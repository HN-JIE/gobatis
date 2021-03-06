# gobatis

目前代码都是基于mysql编写测试的,其他数据库暂时还未做兼容处理

## db数据源配置
### 示例
* db配置示例

db.yml
```yaml
# 数据库配置
db:
  # 数据源名称
  datasource1:
    # 驱动名
    driverName: mysql
    # 数据源
    dataSourceName: root:123456@tcp(127.0.0.1:3306)/test?charset=utf8
    # 连接最大存活时间（单位：s）
    maxLifeTime: 120
    # 最大open连接数
    maxOpenConns: 10
    # 最大挂起连接数
    maxIdleConns: 5
  # 数据源名称
  datasource2:
    # 驱动名
    driverName: mysql
    # 数据源
    dataSourceName: root:123456@tcp(127.0.0.1:3306)/test?charset=utf8
    # 连接最大存活时间（单位：s）
    maxLifeTime: 120
    # 最大open连接数
    maxOpenConns: 10
    # 最大挂起连接数
    maxIdleConns: 5
# 是否显示SQL语句
showSql: true
# 数据表映射文件路径配置
mappers:
  - mapper/userMapper.xml
```

* mapper配置文件示例

mapper/userMapper.xml
```xml
<?xml version="1.0" encoding="utf-8"?>
<mapper namespace="userMapper">
    <select id="findMapById" resultType="Map">
        SELECT id, name FROM user where id=#{id} order by id
    </select>
    <select id="findMapByValue" resultType="Map">
            SELECT id, name FROM user where id=#{0} order by id
        </select>
    <select id="findStructByStruct" resultType="Struct">
        SELECT id, name, crtTm FROM user where id=#{Id} order by id
    </select>
    <insert id="insertStruct">
        insert into user (name, email, crtTm)
        values (#{Name}, #{Email}, #{CrtTm})
    </insert>
    <delete id="deleteById">
        delete from user where id=#{id}
    </delete>
</mapper>
```

## 使用方法
example.go
```go
package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/wenj91/gobatis"
)

type User struct {
	Id    gobatis.NullInt64  `field:"id"`
	Name  gobatis.NullString `field:"name"`
	Email gobatis.NullString `field:"email"`
	CrtTm gobatis.NullTime   `field:"crtTm"`
}

func main() {
	gobatis.ConfInit("db.yml")
	gb := gobatis.NewGoBatis("datasource1")

	//传入id查询Map
	mapRes := make(map[string]interface{})
	err := gb.Select("userMapper.findMapById", map[string]interface{}{"id": 1})(mapRes)
	fmt.Println("userMapper.findMapById-->", mapRes, err)

	//根据传入实体查询对象
	param := User{Id: gobatis.NullInt64{Int64: 1, Valid: true}}
	structRes2 := User{}
	err = gb.Select("userMapper.findStructByStruct", param)(&structRes2)
	fmt.Println("userMapper.findStructByStruct-->", structRes2, err)

	tx, _ := gb.Begin()
	tx.Select("userMapper.findMapById", map[string]interface{}{"id": 1,})(mapRes)
	fmt.Println("tx userMapper.findMapById-->", mapRes, err)
	tx.Commit()
}
```
