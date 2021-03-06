package gobatis

import (
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"strings"
)

func createSqlNode(elems ...element) iSqlNode {
	if len(elems) == 0 {
		return &textSqlNode{""}
	}

	if len(elems) == 1 {
		elem := elems[0]
		if elem.ElementType == eleTpText {
			return &textSqlNode{
				content: elem.Val.(string),
			}
		}

		n := elem.Val.(node)
		if n.Name == "if" {
			sqlNode := createSqlNode(n.Elements...)
			return &ifSqlNode{
				test:    n.Attrs["test"].Value,
				sqlNode: sqlNode,
			}
		}

		if n.Name == "foreach" {
			sqlNode := createSqlNode(n.Elements...)

			open := ""
			openAttr, ok := n.Attrs["open"]
			if ok {
				open = openAttr.Value
			}

			closeStr := ""
			closeAttr, ok := n.Attrs["close"]
			if ok {
				closeStr = closeAttr.Value
			}

			separator := ""
			separatorAttr, ok := n.Attrs["separator"]
			if ok {
				separator = separatorAttr.Value
			}

			itemAttr, ok := n.Attrs["item"]
			if !ok {
				log.Fatalln("No attr:`item` for tag:", n.Name)
			}
			item := itemAttr.Value

			index := ""
			indexAttr, ok := n.Attrs["index"]
			if ok {
				index = indexAttr.Value
			}

			collectionAttr, ok := n.Attrs["collection"]
			if !ok {
				log.Fatalln("No attr:`collection` for tag:", n.Name)
			}
			collection := collectionAttr.Value

			return &foreachSqlNode{
				sqlNode:    sqlNode,
				open:       open,
				close:      closeStr,
				separator:  separator,
				item:       item,
				index:      index,
				collection: collection,
			}
		}

		log.Fatalln("The tag:", n.Name, "not support, current version only support tag:<if> | <foreach>")
	}

	sns := make([]iSqlNode, 0)
	for _, elem := range elems {
		sqlNode := createSqlNode(elem)
		sns = append(sns, sqlNode)
	}

	return &mixedSqlNode{
		sqlNodes: sns,
	}
}

func buildMapperConfig(r io.Reader) *mapperConfig {
	rootNode := parse(r)

	conf := &mapperConfig{
		mappedStmts: make(map[string]*node),
	}

	if rootNode.Name != "mapper" {
		log.Fatalln("Mapper xml must start with `mapper` tag, please check your xml mapperConfig!")
	}

	namespace := ""
	if val, ok := rootNode.Attrs["namespace"]; ok {
		nStr := strings.TrimSpace(val.Value)
		if nStr != "" {
			nStr += "."
		}
		namespace = nStr
	}

	for _, elem := range rootNode.Elements {
		if elem.ElementType == eleTpNode {
			childNode := elem.Val.(node)
			switch childNode.Name {
			case "select", "update", "insert", "delete":
				if childNode.Id == "" {
					log.Fatalln("No id for:", childNode.Name, "Id must be not null, please check your xml mapperConfig!")
				}

				fid := namespace + childNode.Id
				if ok := conf.put(fid, &childNode); !ok {
					log.Fatalln("Repeat id for:", fid, "Please check your xml mapperConfig!")
				}
			}
		}
	}

	return conf
}

func buildDbConfig(ymlStr string) *dbConfig {
	dbconf := &dbConfig{}
	err := yaml.Unmarshal([]byte(ymlStr), &dbconf)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return dbconf
}
