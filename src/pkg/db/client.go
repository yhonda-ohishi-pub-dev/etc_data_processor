package db

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "github.com/yhonda-ohishi-pub-dev/db_service/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ETCMeisaiClient wraps db_service gRPC client for ETC data operations
type ETCMeisaiClient struct {
	conn   *grpc.ClientConn
	client pb.Db_ETCMeisaiServiceClient
}

// NewETCMeisaiClient creates a new gRPC client connecting to db_service
func NewETCMeisaiClient(address string) (*ETCMeisaiClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db_service at %s: %w", address, err)
	}

	return &ETCMeisaiClient{
		conn:   conn,
		client: pb.NewDb_ETCMeisaiServiceClient(conn),
	}, nil
}

// SaveETCData implements handler.DBClient interface
// Converts map data to ETCMeisai proto and saves to database
func (c *ETCMeisaiClient) SaveETCData(data interface{}) error {
	etcData, ok := data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data type, expected map[string]interface{}, got %T", data)
	}

	// Convert map to ETCMeisai proto message
	etcMeisai, err := convertToETCMeisai(etcData)
	if err != nil {
		return fmt.Errorf("failed to convert data: %w", err)
	}

	// Create request using generated proto
	req := &pb.Db_CreateETCMeisaiRequest{
		EtcMeisai: etcMeisai,
	}

	// Call gRPC service
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = c.client.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to save ETC meisai to db_service: %w", err)
	}

	return nil
}

// Close closes the gRPC connection
func (c *ETCMeisaiClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// convertToETCMeisai converts map data to ETCMeisai proto message
func convertToETCMeisai(data map[string]interface{}) (*pb.Db_ETCMeisai, error) {
	// Parse date field
	dateStr, ok := data["date"].(string)
	if !ok || dateStr == "" {
		return nil, fmt.Errorf("missing or invalid 'date' field: got type %T, value %v", data["date"], data["date"])
	}

	// Convert date to RFC3339 format for db_service
	dateToRFC3339, err := formatDateToRFC3339(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to convert date to RFC3339: %w", err)
	}

	// Parse entry_ic
	entryIC, ok := data["entry_ic"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'entry_ic' field")
	}

	// Parse exit_ic
	exitIC, ok := data["exit_ic"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'exit_ic' field")
	}

	// Parse amount
	amount, err := parseInt32(data["amount"])
	if err != nil {
		return nil, fmt.Errorf("invalid 'amount': %w", err)
	}

	// Parse vehicle_type (maps to shashu in db_service)
	vehicleType, err := parseInt32(data["vehicle_type"])
	if err != nil {
		return nil, fmt.Errorf("invalid 'vehicle_type': %w", err)
	}

	// Parse optional card_number (maps to etc_num)
	cardNumber, _ := data["card_number"].(string)

	// ic_fr (入口IC) は optional - *string型
	var icFr *string
	if entryIC != "" {
		icFr = &entryIC
	}

	// Build ETCMeisai proto message
	etcMeisai := &pb.Db_ETCMeisai{
		DateTo:     dateToRFC3339, // RFC3339 format for db_service
		DateToDate: dateStr,       // Keep original date format for date_to_date
		IcFr:       icFr,          // optional: *string型
		IcTo:       exitIC,
		Price:      amount,
		Shashu:     vehicleType,
		EtcNum:     cardNumber,
	}

	// Generate hash for duplicate detection
	etcMeisai.Hash = generateHash(etcMeisai)

	return etcMeisai, nil
}

// parseInt32 converts various types to int32
func parseInt32(value interface{}) (int32, error) {
	switch v := value.(type) {
	case int:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("cannot parse '%s' as int32: %w", v, err)
		}
		return int32(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported type %T for int32 conversion", value)
	}
}

// generateHash generates SHA256 hash for duplicate detection
func generateHash(etcMeisai *pb.Db_ETCMeisai) string {
	icFr := ""
	if etcMeisai.IcFr != nil {
		icFr = *etcMeisai.IcFr
	}
	data := fmt.Sprintf("%s_%s_%s_%d_%s",
		etcMeisai.DateTo,
		icFr,
		etcMeisai.IcTo,
		etcMeisai.Price,
		etcMeisai.EtcNum,
	)
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// formatDateToRFC3339 converts a date string to RFC3339 format
// Input: "2006-01-02" or "2006-01-02T15:04:05" etc.
// Output: "2006-01-02T00:00:00Z" (RFC3339 format)
func formatDateToRFC3339(dateStr string) (string, error) {
	// Already RFC3339 format, return as-is
	if strings.Contains(dateStr, "T") {
		if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
			return t.Format(time.RFC3339), nil
		}
	}

	// Date only format, parse and convert to RFC3339
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("cannot parse date '%s': %w", dateStr, err)
	}

	// Convert to UTC 00:00:00 and format as RFC3339
	return t.UTC().Format(time.RFC3339), nil
}
