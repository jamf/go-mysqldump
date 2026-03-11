package mysqldump

import (
	"bytes"
	"database/sql"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func getMockData() (data *Data, mock sqlmock.Sqlmock, err error) {
	var db *sql.DB
	db, mock, err = sqlmock.New()
	if err != nil {
		return
	}

	mock.ExpectBegin()

	config := mysql.Config{}

	data = &Data{
		Config:     config,
		Connection: db,
	}
	err = data.begin()
	return
}

func c(name string, v interface{}) *sqlmock.Column {
	var t string
	switch reflect.ValueOf(v).Kind() {
	case reflect.String:
		t = "VARCHAR"
	case reflect.Int:
		t = "INT"
	case reflect.Bool:
		t = "BOOL"
	}
	return sqlmock.NewColumn(name).OfType(t, v).Nullable(true)
}

func TestGetTablesOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")

	data.Opts.Schema = "Testdb"

	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("Test_Table_1", "BASE TABLE").
		AddRow("Test_Table_2", "BASE TABLE")

	mock.ExpectQuery("^SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'Testdb' AND table_name LIKE '%%'$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "Test_Table_1", result[0].Name)
	assert.Equal(t, "Test_Table_2", result[1].Name)
}

func TestIgnoreTablesOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")

	data.Opts.Schema = "Testdb"
	data.Opts.IgnoreTables = []string{"Test_Table_1"}

	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("Test_Table_1", "BASE TABLE").
		AddRow("Test_Table_2", "BASE TABLE")

	mock.ExpectQuery("^SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'Testdb' AND table_name LIKE '%%'$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.Equal(t, 1, len(result))
	assert.Equal(t, "Test_Table_2", result[0].Name)
}

func TestGetTablesWithPrefix(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	data.Opts.Schema = "Testdb"
	data.Opts.TablePrefix = "wp_"

	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("wp_posts", "BASE TABLE").
		AddRow("wp_users", "BASE TABLE")

	mock.ExpectQuery("^SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'Testdb' AND table_name LIKE 'wp_%%'$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, 2, len(result))
	assert.Equal(t, "wp_posts", result[0].Name)
}

func TestGetTablesWithSuffix(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	data.Opts.Schema = "Testdb"
	data.Opts.TableSuffix = "_backup"

	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("users_backup", "BASE TABLE")

	mock.ExpectQuery("^SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'Testdb' AND table_name LIKE '%%_backup'$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, 1, len(result))
}

func TestGetTablesWithPrefixAndSuffix(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	data.Opts.Schema = "Testdb"
	data.Opts.TablePrefix = "pre_"
	data.Opts.TableSuffix = "_post"

	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("pre_data_post", "BASE TABLE")

	mock.ExpectQuery("^SELECT table_name, table_type FROM information_schema.tables WHERE table_schema = 'Testdb' AND table_name LIKE 'pre_%%_post'$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	assert.Equal(t, 1, len(result))
	assert.Equal(t, "pre_data_post", result[0].Name)
}

func TestGetTablesComprehensiveFiltering(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	// Настраиваем жесткие фильтры
	data.Opts.Schema = "TargetDB"
	data.Opts.TablePrefix = "prod_"
	data.Opts.IgnoreTables = []string{"prod_secret"}

	// Имитируем ситуацию: в БД есть много таблиц, но SQL-запрос должен ограничить их
	// Мы проверяем, что SQL-запрос содержит правильный WHERE
	rows := sqlmock.NewRows([]string{"table_name", "table_type"}).
		AddRow("prod_users", "BASE TABLE").
		AddRow("prod_secret", "BASE TABLE") // Эту таблицу мы проигнорируем в Go

	// Регулярное выражение проверяет, что мы ищем именно в TargetDB и с нужным префиксом
	mock.ExpectQuery("WHERE table_schema = 'TargetDB' AND table_name LIKE 'prod_%%'").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Должна остаться только одна таблица
	assert.Equal(t, 1, len(result), "Should exclude ignored tables")
	assert.Equal(t, "prod_users", result[0].Name)

	// Проверяем, что 'prod_secret' была отфильтрована кодом Go
	for _, table := range result {
		assert.NotEqual(t, "prod_secret", table.Name)
	}
}

func TestGetTablesNoMatches(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	data.Opts.Schema = "EmptyDB"

	// База ничего не вернула
	rows := sqlmock.NewRows([]string{"table_name", "table_type"})
	mock.ExpectQuery("WHERE table_schema = 'EmptyDB'").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)
	assert.Empty(t, result, "Result should be empty if no tables match criteria")
}

func TestWriteTableView(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	var buf bytes.Buffer
	data.Out = &buf
	assert.NoError(t, data.getTemplates())

	// Mocking a VIEW
	createViewRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("test_view", "CREATE VIEW `test_view` AS SELECT 1")

	mock.ExpectQuery("^SHOW CREATE TABLE `test_view`$").WillReturnRows(createViewRows)

	table := data.createTable("test_view", true)
	err = data.writeTable(table)
	assert.NoError(t, err)

	// Should contain VIEW structure but NO "Dumping data" section
	result := buf.String()
	assert.Contains(t, result, "View structure for view `test_view`")
	assert.NotContains(t, result, "Dumping data for table")
}

func TestVirtualColumnsExclusion(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	// id is regular, full_name is VIRTUAL
	colInfo := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("full_name", "VIRTUAL GENERATED")

	mock.ExpectQuery("^SHOW COLUMNS FROM `test_virtual`$").WillReturnRows(colInfo)

	// Expecting SELECT only for "id"
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("^SELECT `id` FROM `test_virtual`$").WillReturnRows(rows)

	table := data.createTable("test_virtual", false)
	assert.True(t, table.Next())
	assert.Equal(t, []string{"id"}, table.cols)
}

func TestBinaryDataHandling(t *testing.T) {
	data, _, err := getMockData()
	assert.NoError(t, err)

	table := &table{
		data:   data,
		values: []interface{}{&sql.RawBytes{0x00, 0x01, 0x02}},
	}

	result := table.RowValues()
	// Should be prefixed with _binary
	assert.Contains(t, result, "_binary")
}

func TestGetServerVersionOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	rows := sqlmock.NewRows([]string{"Version()"}).AddRow("test_version")
	mock.ExpectQuery("^SELECT version()").WillReturnRows(rows)

	meta := metaData{}
	assert.NoError(t, meta.updateServerVersion(data))
	assert.Equal(t, "test_version", meta.ServerVersion)
}

func TestCreateSQLSQLOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE `Test_Table` (id int)")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(rows)

	table := data.createTable("Test_Table", false)
	result, err := table.CreateSQL()
	assert.NoError(t, err)
	assert.Contains(t, result, "CREATE TABLE")
}

func mockTableSelect(mock sqlmock.Sqlmock, name string) {
	cols := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("name", "")

	rows := sqlmock.NewRowsWithColumnDefinition(c("id", 0), c("name", "")).
		AddRow(1, "Test 1")

	mock.ExpectQuery("^SHOW COLUMNS FROM `" + name + "`$").WillReturnRows(cols)
	mock.ExpectQuery("^SELECT (.+) FROM `" + name + "`$").WillReturnRows(rows)
}

func TestCreateTableValuesSteam(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err)

	mockTableSelect(mock, "test")
	data.Opts.MaxAllowedPacket = 4096

	table := data.createTable("test", false)
	s := table.Stream()
	assert.Contains(t, <-s, "INSERT INTO `test` (`id`, `name`) VALUES (1,'Test 1');")
}
