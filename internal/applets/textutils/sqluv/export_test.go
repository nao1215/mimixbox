package sqluv

// Exported aliases for white-box unit tests.

type (
	SourceKindAlias = sourceKind
	FileFormatAlias = fileFormat
	CompressionAlias = compression
)

var (
	ClassifySource      = classifySource
	ValidateSource      = validateSource
	DetectFormatByName  = detectFormatByName
	SplitCompression    = splitCompression
	TableNameFor        = tableNameFor
	IsMutating          = isMutating
	HistoryPath         = historyPath
	ValidateOutputFmt   = validateOutputFormat
	LoadDelimitedForTst = loadDelimited
)

// Re-export unexported kinds/formats/compression for assertions.
const (
	KindDelimited  = kindDelimited
	KindSQLite     = kindSQLite
	KindUnsupported = kindUnsupported

	FormatCSV     = formatCSV
	FormatTSV     = formatTSV
	FormatLTSV    = formatLTSV
	FormatUnknown = formatUnknown

	CompNone  = compNone
	CompGzip  = compGzip
	CompBzip2 = compBzip2
	CompXz    = compXz
	CompZstd  = compZstd
)

// Columns and Rows expose a loaded table for assertions.
func (t *table) Columns() []string { return t.columns }
func (t *table) Rows() [][]string  { return t.rows }
func (t *table) TableName() string { return t.name }
