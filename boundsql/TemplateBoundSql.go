package boundsql

import (
	"github.com/wenj91/gobatis/mapperstmt"
	"text/template"
	"bytes"
)

func TemplateGetBoundSql(sqlNode mapperstmt.SqlNode, param interface{}) (boundSql BoundSql, err error) {
	eles := sqlNode.Elements
	sqlStr := ""
	for i:=0; i<len(eles); i++ {
		sqlStr += (eles[i].Val.(string))
	}

	t := template.Must(template.New("sql").Parse(sqlStr))

	buf := bytes.NewBufferString("")
	err = t.Execute(buf, param)
	if nil != err {
		return
	}

	boundSql.Sql = buf.String()
	boundSql.ResultType = sqlNode.ResultType

	return
}
