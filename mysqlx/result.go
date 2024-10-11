package mysqlx

type row struct {
	Values map[string]string
}
type result struct {
	Fileds []string
	Rows   []row
}
