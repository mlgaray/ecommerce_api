package mock

// type DataBaseConnection struct {
//	DB *sql.DB
//}
//
// func (m *DataBaseConnection) Connect() (*sql.DB, sqlmock.Sqlmock) {
//	sqlDB, mock, _ := sqlmock.New()
//
//	/*			AddRow(1, "Usuario1", "usuario1@example.com").
//		AddRow(2, "Usuario2", "usuario2@example.com").
//		AddRow(3, "Usuario3", "usuario3@example.com")
//	mock.ExpectQuery("SELECT id, name, email FROM users").WillReturnRows(rows)
//
//	m.DB = sqlDB
//	return m.DB*/
//	return sqlDB, mock
//}
//
//func NewDataBaseConnection() *DataBaseConnection {
//	return &DataBaseConnection{}
//}
