# ETC Data Processor

ETCデータの処理とバリデーションを行うGoライブラリ

## 概要

このプロジェクトは、ETC（Electronic Toll Collection）の利用明細CSVファイルを解析し、データの処理・バリデーションを行うためのライブラリです。

## 特徴

- ✅ **100%テストカバレッジ**（手書きコード）
- 🔧 **包括的なエラーハンドリング**
- 📊 **詳細なカバレッジレポート**
- 🎨 **色付きテスト結果表示**
- 🚀 **高性能なCSV処理**

## 主要機能

### CSVパーサー
- ETC明細CSVファイルの解析
- ヘッダー付き/なしの両方に対応
- 様々な日付フォーマットサポート
- 車種・料金データの正確な処理

### バリデーション
- CSVデータの完全性チェック
- 必須フィールドの検証
- 重複データの検出
- エラーレポート生成

### サービス層
- gRPCサービスインターフェース
- ファイル処理API
- データ変換機能

## アーキテクチャ

```
src/
├── pkg/
│   ├── handler/     # サービス層とバリデーション
│   └── parser/      # CSVパーサー
├── proto/           # プロトコルバッファ定義
├── cmd/server/      # gRPCサーバー
└── internal/        # 内部パッケージ
```

## テスト戦略

### 関数抽出による徹底テスト
- `ParseVehicleClass`: strconv.Atoiエラーパステスト
- `parseDate`: 日付解析の全エッジケース
- `ValidateRecordsAvailable`: データ存在チェック
- `ProcessRecords`: バリデーションエラー処理

### カバレッジ計測
```bash
./show_coverage.sh
```

### テスト実行
```bash
go test ./tests/...
```

## 環境変数

以下の環境変数でサービスの動作を制御できます：

| 変数名 | 説明 | デフォルト値 | 例 |
|--------|------|-------------|-----|
| `GRPC_PORT` | gRPCサーバーのポート番号（優先） | 50051 | `50052` |
| `ETC_PROCESSOR_PORT` | gRPCサーバーのポート番号 | 50051 | `50052` |
| `ETC_PROCESSOR_DB_ADDR` | データベースサービスのアドレス | - | `localhost:50051` |
| `SKIP_DUPLICATES` | 重複チェックの有効/無効 | `true` | `false`, `0` |
| `CSV_BASE_PATH` | CSVファイルのベースパス（最新フォルダ自動検索） | - | `/data/csv` |

### 使用例

```bash
# ポート番号を変更
export GRPC_PORT=50052

# 重複チェックを無効化
export SKIP_DUPLICATES=false

# データベース接続先を指定
export ETC_PROCESSOR_DB_ADDR=localhost:50051

# CSVファイルの自動検索を有効化
# ベースパス内の最新フォルダから自動的にCSVファイルを探します
export CSV_BASE_PATH=/data/csv_files

# サーバー起動
./etc_data_processor
```

#### CSV_BASE_PATH の動作

`CSV_BASE_PATH`を設定すると、以下の動作になります：

1. **ベースパス内のフォルダを検索**: `/data/csv_files/` 内の全フォルダをスキャン
2. **最新フォルダを特定**: フォルダ名でソート（降順）して最新のフォルダを選択
3. **CSVファイルを検索**: 最新フォルダ内の `.csv` ファイルを自動検出
4. **自動処理**: 見つかったCSVファイルを処理

例：
```
/data/csv_files/
  ├── 20251118_120000/
  │   └── etc_data.csv
  └── 20251119_150000/  ← 最新フォルダ（自動選択）
      └── etc_data.csv  ← このファイルが処理される
```

**注意**: `CSV_BASE_PATH`が設定されている場合、リクエストの`csv_file_path`パラメータは無視され、自動検索が優先されます。

## API仕様

### リクエストパラメータ

#### ProcessCSVFile / ProcessCSVData

| パラメータ | 型 | 必須 | デフォルト | 説明 |
|-----------|-----|------|-----------|------|
| `csv_file_path` | string | ✅ | - | CSVファイルのパス（ProcessCSVFileのみ） |
| `csv_data` | string | ✅ | - | CSV文字列データ（ProcessCSVDataのみ） |
| `account_id` | string | ❌ | - | アカウントID（3文字以上、将来のマルチテナント対応用） |
| `skip_duplicates` | bool | ❌ | `true` | 重複チェック（環境変数`SKIP_DUPLICATES`で制御可能） |

**注**: `account_id`はオプショナルです。空文字列を指定するか省略できます。

## 使用技術

- **言語**: Go 1.21+
- **プロトコル**: gRPC
- **テスト**: Go標準テストパッケージ
- **カバレッジ**: go tool cover

## 開発者向け

### セットアップ
```bash
git clone <repository>
cd etc_data_processor
go mod download
```

### テスト実行
```bash
# 全テスト実行
go test ./tests/...

# カバレッジ付きテスト
./show_coverage.sh
```

### ビルド
```bash
go build ./src/cmd/server
```

## カバレッジレポート

現在のテストカバレッジ: **100.0%**（手書きコード）

- ✅ 34関数 - 完全テスト済み
- 🔶 0関数 - 部分的カバレッジ
- ⚠️ 0関数 - 未テスト

## ライセンス

MIT License

## 貢献

プルリクエストを歓迎します。大きな変更を行う場合は、まずissueを作成して変更内容について議論してください。