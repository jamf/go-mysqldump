package mysqldump

import (
	"bytes"
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func getMockData() (data *Data, mock sqlmock.Sqlmock, err error) {
	var db *sql.DB
	db, mock, err = sqlmock.New()
	if err != nil {
		return
	}
	mock.ExpectBegin()

	data = &Data{
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
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Tables_in_Testdb", "Table_type"}).
		AddRow("Test_Table_1", "BASE TABLE").
		AddRow("Test_Table_2", "BASE TABLE")

	mock.ExpectQuery("^SHOW FULL TABLES$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.EqualValues(t, []string{"Test_Table_1", "Test_Table_2"}, tableNames(result))
}

func TestIgnoreTablesOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Tables_in_Testdb", "Table_type"}).
		AddRow("Test_Table_1", "BASE TABLE").
		AddRow("Test_Table_2", "BASE TABLE")

	mock.ExpectQuery("^SHOW FULL TABLES$").WillReturnRows(rows)

	data.IgnoreTables = []string{"Test_Table_1"}

	result, err := data.getTables()
	assert.NoError(t, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.EqualValues(t, []string{"Test_Table_2"}, tableNames(result))
}

func TestGetTablesNil(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Tables_in_Testdb", "Table_type"}).
		AddRow("Test_Table_1", "BASE TABLE").
		AddRow(nil, "BASE TABLE").
		AddRow("Test_Table_3", "BASE TABLE")

	mock.ExpectQuery("^SHOW FULL TABLES$").WillReturnRows(rows)

	result, err := data.getTables()
	assert.NoError(t, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.EqualValues(t, []string{"Test_Table_1", "Test_Table_3"}, tableNames(result))
}

func tableNames(tables []*table) []string {
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, t.Name)
	}
	return names
}

func TestGetServerVersionOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Version()"}).
		AddRow("test_version")

	mock.ExpectQuery("^SELECT version()").WillReturnRows(rows)

	meta := metaData{}

	assert.NoError(t, meta.updateServerVersion(data), "error was not expected while updating stats")

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.Equal(t, "test_version", meta.ServerVersion)
}

func TestCreateSQLSQLOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(rows)

	table := data.createTable("Test_Table", false)

	result, err := table.CreateSQL()
	assert.NoError(t, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1"

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateSQLQueryFail(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	queryError := errors.New("query failure")
	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnError(queryError)

	table := data.createTable("Test_Table", false)

	result, err := table.CreateSQL()
	assert.Error(t, err)
	assert.Equal(t, queryError, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := ""

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateSQLWrongTable(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Diff_Table", "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(rows)

	table := data.createTable("Test_Table", false)

	result, err := table.CreateSQL()
	assert.Error(t, err)
	expectedError := errors.New("returned table is not the same as requested table")
	assert.Equal(t, expectedError, err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := ""

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func TestCreateTableInvalidColumns(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	rows := sqlmock.NewRows([]string{"Table"}).
		AddRow("Test_Table")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(rows)

	table := data.createTable("Test_Table", false)

	result, err := table.CreateSQL()
	assert.Error(t, err)
	assert.Equal(t, errors.New("database column information is malformed"), err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := ""

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("expected %#v, got %#v", expectedResult, result)
	}
}

func mockTableSelect(mock sqlmock.Sqlmock, name string) {
	cols := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("email", "").
		AddRow("name", "")

	rows := sqlmock.NewRowsWithColumnDefinition(c("id", 0), c("email", ""), c("name", "")).
		AddRow(1, "test@test.de", "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SHOW COLUMNS FROM `" + name + "`$").WillReturnRows(cols)
	mock.ExpectQuery("^SELECT (.+) FROM `" + name + "`$").WillReturnRows(rows)
}

func TestCreateTableRowValues(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	mockTableSelect(mock, "test")

	table := data.createTable("test", false)

	assert.True(t, table.Next())

	result := table.RowValues()
	assert.NoError(t, table.Err)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	assert.EqualValues(t, "(1,'test@test.de','Test Name 1')", result)
}

func TestCreateTableValuesSteam(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	mockTableSelect(mock, "test")

	data.MaxAllowedPacket = 4096

	table := data.createTable("test", false)

	s := table.Stream()
	assert.EqualValues(t, "INSERT INTO `test` (`id`, `email`, `name`) VALUES (1,'test@test.de','Test Name 1'),(2,'test2@test.de','Test Name 2');", <-s)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")
}

func TestCreateTableValuesSteamSmallPackets(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	mockTableSelect(mock, "test")

	data.MaxAllowedPacket = 64

	table := data.createTable("test", false)

	s := table.Stream()
	assert.EqualValues(t, "INSERT INTO `test` (`id`, `email`, `name`) VALUES (1,'test@test.de','Test Name 1');", <-s)
	assert.EqualValues(t, "INSERT INTO `test` (`id`, `email`, `name`) VALUES (2,'test2@test.de','Test Name 2');", <-s)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")
}

func TestCreateTableAllValuesWithNil(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	cols := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("email", "").
		AddRow("name", "")

	rows := sqlmock.NewRowsWithColumnDefinition(c("id", 0), c("email", ""), c("name", "")).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2").
		AddRow(3, "", "Test Name 3")

	mock.ExpectQuery("^SHOW COLUMNS FROM `test`$").WillReturnRows(cols)
	mock.ExpectQuery("^SELECT (.+) FROM `test`$").WillReturnRows(rows)

	table := data.createTable("test", false)

	results := make([]string, 0)
	for table.Next() {
		row := table.RowValues()
		assert.NoError(t, table.Err)
		results = append(results, row)
	}

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResults := []string{"(1,NULL,'Test Name 1')", "(2,'test2@test.de','Test Name 2')", "(3,'','Test Name 3')"}

	assert.EqualValues(t, expectedResults, results)
}

func TestCreateTableEmptyValuesByColumnType(t *testing.T) {
	tests := []struct {
		dbType   string
		value    string
		expected []string
	}{
		{"CHAR", "it's", []string{"('')", "(NULL)", "('it\\'s')", "('')"}},
		{"VARCHAR", "it's", []string{"('')", "(NULL)", "('it\\'s')", "('')"}},
		{"TEXT", "a\\b", []string{"('')", "(NULL)", "('a\\\\b')", "('')"}},
		{"MEDIUMTEXT", "x", []string{"('')", "(NULL)", "('x')", "('')"}},
		{"ENUM", "x", []string{"('')", "(NULL)", "('x')", "('')"}},
		{"SET", "x", []string{"('')", "(NULL)", "('x')", "('')"}},
		{"VARBINARY", "x", []string{"('')", "(NULL)", "(_binary 'x')", "('')"}},
		{"BLOB", "it's", []string{"('')", "(NULL)", "(_binary 'it\\'s')", "('')"}},
		{"GEOMETRY", "x", []string{"('')", "(NULL)", "(_binary 'x')", "('')"}},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			data, mock, err := getMockData()
			assert.NoError(t, err, "an error was not expected when opening a stub database connection")
			defer data.Close()

			cols := sqlmock.NewRows([]string{"Field", "Extra"}).
				AddRow("v", "")

			// go-sql-driver reports sql.RawBytes as the scan type of every string-family column.
			col := sqlmock.NewColumn("v").OfType(tt.dbType, sql.RawBytes{}).Nullable(true)
			rows := sqlmock.NewRowsWithColumnDefinition(col).
				AddRow("").
				AddRow(nil).
				AddRow(tt.value).
				AddRow("")

			mock.ExpectQuery("^SHOW COLUMNS FROM `test`$").WillReturnRows(cols)
			mock.ExpectQuery("^SELECT (.+) FROM `test`$").WillReturnRows(rows)

			table := data.createTable("test", false)

			results := make([]string, 0)
			for table.Next() {
				results = append(results, table.RowValues())
			}
			assert.NoError(t, table.Err)
			assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")
			assert.EqualValues(t, tt.expected, results)
		})
	}
}

func TestNullBytesScanCopiesSource(t *testing.T) {
	var n nullBytes
	src := []byte("first")
	assert.NoError(t, n.Scan(src))
	copy(src, "xxxxx")
	assert.True(t, n.Valid)
	assert.Equal(t, "first", string(n.Bytes))

	assert.NoError(t, n.Scan(nil))
	assert.False(t, n.Valid)

	assert.NoError(t, n.Scan([]byte{}))
	assert.True(t, n.Valid)
	assert.Empty(t, n.Bytes)

	assert.Error(t, n.Scan(1))
}

func TestCreateTableOk(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	createTableRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	createTableValueCols := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("email", "").
		AddRow("name", "")

	createTableValueRows := sqlmock.NewRowsWithColumnDefinition(c("id", 0), c("email", ""), c("name", "")).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SHOW COLUMNS FROM `Test_Table`$").WillReturnRows(createTableValueCols)
	mock.ExpectQuery("^SELECT (.+) FROM `Test_Table`$").WillReturnRows(createTableValueRows)

	var buf bytes.Buffer
	data.Out = &buf
	data.MaxAllowedPacket = 4096

	assert.NoError(t, data.getTemplates())

	table := data.createTable("Test_Table", false)

	data.writeTable(table)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := `
--
-- Table structure for table ~Test_Table~
--

DROP TABLE IF EXISTS ~Test_Table~;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE 'Test_Table' (~id~ int(11) NOT NULL AUTO_INCREMENT,~s~ char(60) DEFAULT NULL, PRIMARY KEY (~id~))ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table ~Test_Table~
--

LOCK TABLES ~Test_Table~ WRITE;
/*!40000 ALTER TABLE ~Test_Table~ DISABLE KEYS */;
INSERT INTO ~Test_Table~ (~id~, ~email~, ~name~) VALUES (1,NULL,'Test Name 1'),(2,'test2@test.de','Test Name 2');
/*!40000 ALTER TABLE ~Test_Table~ ENABLE KEYS */;
UNLOCK TABLES;
`
	result := strings.Replace(buf.String(), "`", "~", -1)
	assert.Equal(t, expectedResult, result)
}

func TestCreateTableOkSmallPackets(t *testing.T) {
	data, mock, err := getMockData()
	assert.NoError(t, err, "an error was not expected when opening a stub database connection")
	defer data.Close()

	createTableRows := sqlmock.NewRows([]string{"Table", "Create Table"}).
		AddRow("Test_Table", "CREATE TABLE 'Test_Table' (`id` int(11) NOT NULL AUTO_INCREMENT,`s` char(60) DEFAULT NULL, PRIMARY KEY (`id`))ENGINE=InnoDB DEFAULT CHARSET=latin1")

	createTableValueCols := sqlmock.NewRows([]string{"Field", "Extra"}).
		AddRow("id", "").
		AddRow("email", "").
		AddRow("name", "")

	createTableValueRows := sqlmock.NewRowsWithColumnDefinition(c("id", 0), c("email", ""), c("name", "")).
		AddRow(1, nil, "Test Name 1").
		AddRow(2, "test2@test.de", "Test Name 2")

	mock.ExpectQuery("^SHOW CREATE TABLE `Test_Table`$").WillReturnRows(createTableRows)
	mock.ExpectQuery("^SHOW COLUMNS FROM `Test_Table`$").WillReturnRows(createTableValueCols)
	mock.ExpectQuery("^SELECT (.+) FROM `Test_Table`$").WillReturnRows(createTableValueRows)

	var buf bytes.Buffer
	data.Out = &buf
	data.MaxAllowedPacket = 64

	assert.NoError(t, data.getTemplates())

	table := data.createTable("Test_Table", false)

	data.writeTable(table)

	// we make sure that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expections")

	expectedResult := `
--
-- Table structure for table ~Test_Table~
--

DROP TABLE IF EXISTS ~Test_Table~;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
 SET character_set_client = utf8mb4 ;
CREATE TABLE 'Test_Table' (~id~ int(11) NOT NULL AUTO_INCREMENT,~s~ char(60) DEFAULT NULL, PRIMARY KEY (~id~))ENGINE=InnoDB DEFAULT CHARSET=latin1;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table ~Test_Table~
--

LOCK TABLES ~Test_Table~ WRITE;
/*!40000 ALTER TABLE ~Test_Table~ DISABLE KEYS */;
INSERT INTO ~Test_Table~ (~id~, ~email~, ~name~) VALUES (1,NULL,'Test Name 1');
INSERT INTO ~Test_Table~ (~id~, ~email~, ~name~) VALUES (2,'test2@test.de','Test Name 2');
/*!40000 ALTER TABLE ~Test_Table~ ENABLE KEYS */;
UNLOCK TABLES;
`
	result := strings.Replace(buf.String(), "`", "~", -1)
	assert.Equal(t, expectedResult, result)
}
