package db

import (
	"fmt"
	stdpath "path"
	"strings"
	"log"
	"github.com/jmorganca/ollama/api"
	"github.com/studio-b12/gowebdav"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"io/ioutil"
	"context"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"io"
	"database/sql"
)

func whereInParent(parent string) *gorm.DB {
	if parent == "/" {
		return db.Where("1 = 1")
	}
	return db.Where(fmt.Sprintf("%s LIKE ?", columnName("parent")),
		fmt.Sprintf("%s/%%", parent)).
		Or(fmt.Sprintf("%s = ?", columnName("parent")), parent)
}

func CreateSearchNode(node *model.SearchNode) error {
	err := db.Create(node).Error
    if err != nil {
        return err
    }

    fmt.Printf("New node Parent: %s\n", node.Parent)
    fmt.Printf("New node Name: %s\n", node.Name)
	llava(node)

    return nil
}
	//notice that [info] has os.FileInfo type

	func llava(node *model.SearchNode) {
		db, err := sql.Open("sqlite3", "./data/data.db")
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	
		// 查询表的列信息
		rows, err := db.Query("PRAGMA table_info(x_search_nodes)")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
	
		// 检查是否存在description列
		var hasDescription bool
		for rows.Next() {
			var cid int
			var name string
			var dtype string
			var notnull int
			var dflt_value interface{}
			var pk int
			if err := rows.Scan(&cid, &name, &dtype, &notnull, &dflt_value, &pk); err != nil {
				log.Fatal(err)
			}
			if name == "description" {
				hasDescription = true
				break
			}
		}
	
		// 如果不存在description列，则添加
		if !hasDescription {
			_, err = db.Exec("ALTER TABLE x_search_nodes ADD COLUMN description TEXT")
			if err != nil {
				log.Fatal(err)
			}
		}else {
			fmt.Println("Column 'description' already exists.")
		}
	
		client, err := api.ClientFromEnvironment()
		if err != nil {
			fmt.Printf("%+v", err)
			log.Fatal(err)
		}
	
		if err != nil {
			fmt.Printf("%+v", err)
			log.Fatal(err)
		}
		root := "http://192.168.10.88:5244/dav"
		user := "admin"
		password := "admin"
	
		c := gowebdav.NewClient(root, user, password)
		c.Connect()
	
		webdavFilePath := "/" + node.Parent + "/" + node.Name
		reader, err := c.ReadStream(webdavFilePath)
		if err != nil {
			fmt.Printf("%+v", err)
			log.Fatal(err)
		}

		if reader == nil {
			log.Fatal("Stream is nil")
		}

		data, err := ioutil.ReadAll(reader)
		info, err := c.Stat(webdavFilePath)
		if err != nil {
			fmt.Printf("%+v", err)
			log.Fatal(err)
		}

		if info == nil {
			log.Fatal("FileInfo is nil")
		}
		var description sql.NullString
		err = db.QueryRow("SELECT description FROM x_search_nodes WHERE name = ?", node.Name).Scan(&description)
		if err != nil && err != sql.ErrNoRows {
			fmt.Printf("%+v", err)
			log.Fatal(err)
		}
		// 检查description是否为空
		if description.Valid {
			fmt.Println("Description:", description.String)
		 } else {
			fmt.Println("Description is NULL")
			infoStr := fmt.Sprintf("%v", info)
			if strings.Contains(infoStr, "image") {
				fmt.Println("CTYPE contains 'image'")
				req := &api.GenerateRequest{
					Model:  "llava",
					Prompt: "Describe ",
					Images: []api.ImageData{data},
					Stream: new(bool),
				}
			
				ctx := context.Background()
				respFunc := func(resp api.GenerateResponse) error {
					file, err := os.Create("response.txt")
					if err != nil {
						fmt.Printf("%+v", err)
						log.Fatal(err)
					}
					defer file.Close()
					_, err = io.WriteString(file, resp.Response)
					if err != nil {
						fmt.Printf("%+v", err)
						log.Fatal(err)
					}
			
					// 确保所有的写入操作都已经完成
					file.Sync()
					response:=resp.Response
					_, err = db.Exec("UPDATE x_search_nodes SET description = ? WHERE name = ?", resp.Response, node.Name)
					if err != nil {
						fmt.Printf("%+v", err)
						log.Fatal(err)
					}
					
					fmt.Print(resp.Response)
					if response == "" {
						fmt.Println("resp.Response is empty")
						return nil
					}
					return nil
				}
				
			
				err = client.Generate(ctx, req, respFunc)
				if err != nil {
					fmt.Printf("%+v", err)
					log.Fatal(err)
				}
				fmt.Println()
			} else {
				fmt.Println("CTYPE does not contain 'image'")
			}
		}
		//notice that [info] has os.FileInfo type
		
		
	}
	

func BatchCreateSearchNodes(nodes *[]model.SearchNode) error {
	return db.CreateInBatches(nodes, 1000).Error
}

func DeleteSearchNodesByParent(path string) error {
	path = utils.FixAndCleanPath(path)
	err := db.Where(whereInParent(path)).Delete(&model.SearchNode{}).Error
	if err != nil {
		return err
	}
	dir, name := stdpath.Split(path)
	return db.Where(fmt.Sprintf("%s = ? AND %s = ?",
		columnName("parent"), columnName("name")),
		dir, name).Delete(&model.SearchNode{}).Error
}

func ClearSearchNodes() error {
	return db.Where("1 = 1").Delete(&model.SearchNode{}).Error
}

func GetSearchNodesByParent(parent string) ([]model.SearchNode, error) {
	var nodes []model.SearchNode
	if err := db.Where(fmt.Sprintf("%s = ?",
		columnName("parent")), parent).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func SearchNode(req model.SearchReq, useFullText bool) ([]model.SearchNode, int64, error) {
	var searchDB *gorm.DB
	if !useFullText || conf.Conf.Database.Type == "sqlite3" {
		keywordsClause := db.Where("1 = 1")
		for _, keyword := range strings.Fields(req.Keywords) {
			keywordsClause = keywordsClause.Where("description LIKE ?", fmt.Sprintf("%%%s%%", keyword))
		}
		searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).Where(keywordsClause)
	} else {
		switch conf.Conf.Database.Type {
		case "mysql":
			searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).
				Where("MATCH (name) AGAINST (? IN BOOLEAN MODE)", "'*"+req.Keywords+"*'")
		case "postgres":
			searchDB = db.Model(&model.SearchNode{}).Where(whereInParent(req.Parent)).
				Where("to_tsvector(name) @@ to_tsquery(?)", strings.Join(strings.Fields(req.Keywords), " & "))
		}
	}

	if req.Scope != 0 {
		isDir := req.Scope == 1
		searchDB.Where(db.Where("is_dir = ?", isDir))
	}

	var count int64
	if err := searchDB.Count(&count).Error; err != nil {
		return nil, 0, errors.Wrapf(err, "failed get search items count")
	}
	var files []model.SearchNode
	if err := searchDB.Order("name asc").Offset((req.Page - 1) * req.PerPage).Limit(req.PerPage).
		Find(&files).Error; err != nil {
		return nil, 0, err
	}
	return files, count, nil
}
