# EasyJSONGo
一个十分容易使用的Go语言JSON库（解析JSON、生成JSON）
EasyJSON的[Java版本](https://github.com/373518155/EasyJSON)

## 开始使用

### 引用库文件
使用go get下载
> go get github.com/373518155/EasyJSONGo


然后在代码中引入库
> import "github.com/373518155/EasyJSONGo"


### 使用示例，一行代码用过一百句废话
```go
package main

import (
	"github.com/373518155/EasyJSONGo"
	"fmt"
)

func main()  {
	/*************************************
	解析JSON字符串
	 *************************************/
	jsonStr := `{
	"name": "Go in Action",
	"price": 138.27,
	"authors": ["Raoul-Gabriel Urma", "Mario Fusco", "Alan Mycroft"],
	"pages": 550,
	"available": true,
	"publisher": null,
	"chapters": [{
			"title": "Introduction",
			"pages": 22
		},
		{
			"title": "Basic Go",
			"pages": 33
		},
		{
			"title": "Advanced Go",
			"pages": 44
		}
	]
}`

	// 解析图书信息的JSON字符串，得到easyJSON对象
	easyJSON, _ := EasyJSON.Parse(jsonStr)
	// 获取书名
	name, _ := easyJSON.GetString("name")
	fmt.Println(name)  // 输出: Go in Action

	// 获取第3个作者
	author3, _ := easyJSON.GetString("authors[2]")
	fmt.Println(author3)  // 输出: Alan Mycroft

	// 获取第2章的标题和页码
	chapter2Title, _ := easyJSON.GetString("chapters[1].title")
	fmt.Println(chapter2Title) // 输出: Basic Go
	chapter2Pages, _ := easyJSON.GetInt64("chapters[1].pages")
	fmt.Println(chapter2Pages) // 输出: 33

	/*************************************
	生成JSON字符串
	使用EasyJSON.Array()生成数组
	使用EasyJSON.Object()生成对象
	 *************************************/
	bookObject := EasyJSON.Object(
		"name", "Go in Action",
			"price", 138.27,
			"authors", EasyJSON.Array("Raoul-Gabriel Urma", "Mario Fusco", "Alan Mycroft"),
			"chapters", EasyJSON.Array(
				EasyJSON.Object("title", "Introduction", "pages", 22),
				EasyJSON.Object("title", "Basic Go", "pages", 33)),
			"pages", 550)

	fmt.Println(bookObject.String())
	/*
	输出:
	{
		"name": "Go in Action",
		"price": 138.27,
		"authors": ["Raoul-Gabriel Urma", "Mario Fusco", "Alan Mycroft"],
		"chapters": [{
			"title": "Introduction",
			"pages": 22
		}, {
			"title": "Basic Go",
			"pages": 33
		}],
		"pages": 550
	}
	 */
}

```