/**********************
小说采集器
开发：FJ浪淘沙   微信fjfyfh
日期：2019-09-04
***********************/
package main

import (
	"caiji/tool"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"database/sql"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var InsertSql string
var filename string //= time.Now().Format("20060102150405")

func httpGet(url string) (res string, err error) {

	re, err1 := http.Get(url)

	if err1 != nil {
		err = err1
		return
	}

	defer re.Body.Close()

	buf := make([]byte, 1024*4) //定义一个缓存切片

	for {
		n, err2 := re.Body.Read(buf)

		if n == 0 && err2 == io.EOF {
			// fmt.Println("re.Body.Read err:", err2)
			// return
			break
		}

		res += string(buf[:n])
	}
	return res, err
}

func Dowork(url, title string, page chan string, db *sql.DB, lastID int64) {

	res, err := httpGet(url)
	if err != nil {
		fmt.Printf("httpGet %s \n", err)
		return
	}

	//输出是否有内容
	// fmt.Println(res)

	zhangjie, err := os.Create("zhangjie.txt")
	if err != nil {
		fmt.Printf("os.Create 创建失败 %s \n", err)
		return
	}

	_, err2 := zhangjie.WriteString(res)

	if err2 != nil {
		fmt.Printf("zhangjie.WriteString 写入失败 %s \n", err2)
		return
	}

	defer zhangjie.Close()

	//开始用正则匹配
	// <h1>第267章 你媳妇跟一个男的跑了</h1>
	regx_zhangjie := regexp.MustCompilePOSIX(`<h1>(.*?)</h1>`)
	regx_zhangjie_content := regexp.MustCompilePOSIX(`<div id="content" class="showtxt">(.*?)m\.qb5200\.tw</div>`)
	if regx_zhangjie == nil || regx_zhangjie_content == nil {
		fmt.Printf("regexp.MustCompilePOSIX 错误  \n")
		return
	}

	zhangjieName := regx_zhangjie.FindAllStringSubmatch(res, -1)            //获取章节名称
	zhangjieContent := regx_zhangjie_content.FindAllStringSubmatch(res, -1) //获取章节内容
	//章节名称、章节内容
	zhang_name, zhang_content := zhangjieName[0][1], zhangjieContent[0][1]
	// fmt.Println("章节名称：", zhangjieName[0][1])
	// fmt.Println("章节内容：", zhangjieContent[0][1])

	//判断改文件是否存在
	var zhang *os.File
	var err_ error
	zhang, err_ = os.OpenFile(filename+".txt", os.O_WRONLY, 0644)
	// file_bool := tool.Exists("zhang.txt")
	if err_ != nil && os.IsNotExist(err_) {

		zhang, err_ = os.Create(filename + ".txt")
		if err_ != nil {
			fmt.Printf(" os.Create 章节TXT 错误 %s \n", err_)
			return
		}

	}

	/*	buf := make([]byte, 1024*4)

		for {

			n, err := zhang.Read(buf)
			if err == io.EOF && n == 0 {
				break
			}

		}
	*/

	// fmt.Print(string(buf))

	// 查找文件末尾的偏移量
	n, _ := zhang.Seek(0, 2)

	defer zhang.Close()
	zhang_content = strings.Replace(zhang_content, "<br />", "\n", -1)                         //替换不必要的字符串
	zhang_content = strings.Replace(zhang_content, "("+url+")", "", -1)                        //替换不必要的字符串
	zhang_content = strings.Replace(zhang_content, "&nbsp;", "\t", -1)                         //替换不必要的字符串
	zhang_content = strings.Replace(zhang_content, "<script>chaptererror();</script>", "", -1) //替换不必要的字符串
	// zhang_content = strings.Replace(zhang_content, "请记住本书首发域名：www.qb5200.tw。全本小说网手机版阅读网址：", "", -1)                    //替换不必要的字符串
	// zhang_content = strings.Replace(zhang_content, "全本小说网手机版阅读", "1", -1) //替换不必要的字符串
	zhang_content = strings.Replace(zhang_content, "www.qb5200.tw", "", -1)

	// str := "XBodyContentX"
	// content := str[1 : len(str)-1]

	zhang_content = zhang_content[0 : len(zhang_content)-50] //将内容后面的一句话去掉【注释：无法替换才】

	content_txt := zhang_name + "\n\n" + zhang_content + "\n\n"
	fmt.Printf("content 长度 len=%d \n", len(zhang_content))
	// zhang.WriteString("\n")
	// zhang.WriteString(zhang_name)
	// zhang.WriteString("\n\n")
	// zhang.WriteString(zhang_content)
	// zhang.WriteString("\n")

	// 从末尾的偏移量开始写入内容
	_, err = zhang.WriteAt([]byte(content_txt), n)

	// time_at := strconv.Itoa(time.Now().Year()) + "-" + time.Now().Month().String() + "-" + strconv.Itoa(time.Now().Day()) + " " + strconv.Itoa(time.Now().Hour()) + ":" + strconv.Itoa(time.Now().Minute()) + ":" + strconv.Itoa(time.Now().Second())
	time_at := time.Now().Format("2006-01-02 15:04:05")
	InsertSql += "insert into n_book_content(title,content,b_id,create_time) values('" + zhang_name + "','" + zhang_content + "',6,'" + time_at + "');"

	// insert
	stmt, err := db.Prepare("insert into n_book_content(title,content,b_id,create_time) values(?,?,?,?)")
	checkErr(err)
	zhang_name = tool.ConvertTostring(zhang_name, "gbk", "utf8")
	zhang_content = tool.ConvertTostring(zhang_content, "gbk", "utf8")

	ress, err_sql := stmt.Exec(zhang_name, zhang_content, lastID, time_at)
	checkErr(err_sql)

	_, err_lastid := ress.LastInsertId()
	if err_lastid != nil {
		fmt.Printf(" insert into n_book_content(title,content,b_id,create_time) values(?,?,?,?) res.LastInsertId %s \n", err_lastid)
		return
	}

	page <- url

}

//检验错误
func checkErr(err error) {

	if err != nil {
		panic(err)
	}
}

func main() {

	//链接数据库
	db, err := sql.Open("mysql", "root:root@tcp(192.168.2.239:3306)/novel?charset=utf8")
	if err != nil {
		fmt.Printf("sql.Open 链接错误 %s \n", err)
		return
	}

	var (
		BookUrl       string
		ClassifBookID int
	)

	fmt.Println("请输入小说书地址：")
	fmt.Scan(&BookUrl)

	fmt.Println("请输入分类小说ID：")
	fmt.Scan(&ClassifBookID)

	page := make(chan string)

	url := strings.TrimSpace(BookUrl) // "https://www.qb5200.tw/xiaoshuo/72/72530/" //https://www.qb5200.tw/xiaoshuo/69/69175/  https://www.qb5200.tw/xiaoshuo/60/60400/

	res, err := http.Get(url)

	if err != nil {
		fmt.Printf("http.get 获取错误 #%v# ", err)
		return
	}

	defer res.Body.Close()

	fmt.Printf("Status = %s \n ", res.Status)
	// fmt.Printf("StatusCode  = %d \n ", res.StatusCode)
	// fmt.Printf("Proto = %s \n ", res.Proto)
	// fmt.Printf("Header  = %s \n ", res.Header)
	// fmt.Printf(" Body  = %s \n ", res.Body)

	buf := make([]byte, 1024*4)
	var tmp string
	for {
		n, err := io.ReadFull(res.Body, buf)

		if n == 0 && err != nil {
			fmt.Printf("程序已经读完缓存 io.ReadFull %s \n", err)
			break
		}

		// fmt.Printf("buf= #%s#", buf)

		tmp += string(buf[:n])

	}

	//写入文本
	file, err := os.Create("demo.txt")
	if err != nil {
		fmt.Printf("os.Create 出现错误 %s \n", err)
		return
	}
	_, err1 := file.WriteString(tmp)
	if err1 != nil {
		fmt.Printf("file.Write 出现问题 %s \n ", err1)
		return
	}
	// fmt.Printf("tmp = #%s# \n", tmp) //打印出来
	defer file.Close()

	// var title, author, book_name, novel_status, description, image string

	//正在过滤
	regx_title := regexp.MustCompilePOSIX(`<meta property="og:title" content="(.*?)"/>`)                                       //标题正则匹配
	regx_author := regexp.MustCompilePOSIX(`<meta property="og:novel:author" content="(.*?)"/>`)                               //作者
	regx_book_name := regexp.MustCompilePOSIX(`<meta property="og:novel:book_name" content="(.*?)"/>`)                         //书名称
	regx_novel_status := regexp.MustCompilePOSIX(`<meta property="og:novel:status" content="(.*?)"/>`)                         //连载
	regx_description := regexp.MustCompilePOSIX(`<meta property="og:description" content="([[:space:]]*?.*?[[:space:]]*?)"/>`) //描述 ([[:space:]]+.*?)  / ([[:space:]]*?.*?[[:space:]]*?)
	regx_image := regexp.MustCompilePOSIX(`<meta property="og:image" content="(.*?)"/>`)                                       //图片

	fmt.Println("当前的 regx_title = ", regx_title)
	fmt.Println("当前的 regx_author = ", regx_author)
	fmt.Println("当前的 regx_book_name = ", regx_book_name)
	fmt.Println("当前的 regx_novel_status = ", regx_novel_status)
	fmt.Println("当前的 regx_description = ", regx_description)
	fmt.Println("当前的 regx_image = ", regx_image)

	// fmt.Printf("返回来的正则字符串= %s \n", regx_title.String())

	var jieguo_title [][]string = regx_title.FindAllStringSubmatch(tmp, -1)
	var jieguo_author [][]string = regx_author.FindAllStringSubmatch(tmp, -1)
	var jieguo_book_name [][]string = regx_book_name.FindAllStringSubmatch(tmp, -1)
	var jieguo_novel_status [][]string = regx_novel_status.FindAllStringSubmatch(tmp, -1)
	var jieguo_description [][]string = regx_description.FindAllStringSubmatch(tmp, -1)
	var jieguo_image [][]string = regx_image.FindAllStringSubmatch(tmp, -1)

	// fmt.Println("this is jieguo_description : ", jieguo_description)

	// fmt.Println("len == ", len(jieguo_title))

	fmt.Println(" 查看当前的值 jieguo_title= ", jieguo_title[0][1]) // 因为这里只有一个标题（指的是这个正则）
	fmt.Println(" 查看当前的值 jieguo_author= ", jieguo_author[0][1])
	fmt.Println(" 查看当前的值 jieguo_book_name= ", jieguo_book_name[0][1])
	fmt.Println(" 查看当前的值 jieguo_novel_status= ", jieguo_novel_status[0][1])
	fmt.Println(" 查看当前的值 jieguo_description= ", jieguo_description[0][1])
	fmt.Println(" 查看当前的值 jieguo_image= ", jieguo_image[0][1])
	author, book_name, novel_status, description, image := jieguo_author[0][1], jieguo_book_name[0][1], jieguo_novel_status[0][1], jieguo_description[0][1], jieguo_image[0][1]

	// <dd><a href ="/xiaoshuo/69/69175/103886903.html">第267章 你媳妇跟一个男的跑了</a></dd>
	regx_url := regexp.MustCompilePOSIX(`<dd><a href ="(.*?)">(.*?)</a></dd>`) //正则匹配章节的URL
	var jieguo_url [][]string = regx_url.FindAllStringSubmatch(tmp, -1)

	zhang_url := make([]string, 0)
	// fmt.Println(jieguo_url)
	for _, data := range jieguo_url {

		// title = data[1]
		zhang_url = append(zhang_url, "https://www.qb5200.tw"+data[1])
		// fmt.Println("匹配的项 =", "https://www.qb5200.tw"+data[1]) //如果是data[1]代表的是正则表达式中第一个小括号 data[2]表示是正则表达式中第二个

	}
	//切片去重
	// zhang_new_url := tool.RemoveRepeatedElement(zhang_url)
	zhang_new_url := zhang_url[12:len(zhang_url)]
	fmt.Printf("当前章节长度 len = %d  \n", len(zhang_url))
	fmt.Printf("新的章节长度 len = %d  \n", len(zhang_new_url))
	time_at := time.Now().Format("2006-01-02 15:04:05")
	//将书的信息插入到数据库中，并且返回数据库中书的id
	// insert  	bookSql := "insert into n_book(title,img,author,`describe`,content,c_id,create_time,book_status) values('" + book_name + "','" + image + "','" + author + "','" + description + "','" + description + "',16,'" + time_at + "','" + novel_status + "');"

	stmt, err := db.Prepare("insert into n_book(title,img,author,`describe`,content,c_id,create_time,book_status) values(?,?,?,?,?,?,?,?)")
	checkErr(err)

	book_name = tool.ConvertTostring(book_name, "gbk", "utf8")
	image = tool.ConvertTostring(image, "gbk", "utf8")
	author = tool.ConvertTostring(author, "gbk", "utf8")
	description = tool.ConvertTostring(description, "gbk", "utf8")
	novel_status = tool.ConvertTostring(novel_status, "gbk", "utf8")

	// fmt.Printf("当前gbk 转码 = %s %s %s %s %s \n", book_name, image, author, description, novel_status)
	// return

	ress, err_sql := stmt.Exec(book_name, image, author, description, description, ClassifBookID, time_at, novel_status)
	fmt.Println("insert into n_book(title,img,author,`describe`,content,c_id,create_time,book_status) values('" + book_name + "','" + image + "','" + author + "','" + description + "','" + description + "',16,'" + time_at + "','" + novel_status + "');")
	checkErr(err_sql)
	lastID, err := ress.LastInsertId() //LastInsertId返回一个数据库生成的回应命令的整数。
	checkErr(err)

	filename = tool.FontToPinyin(book_name) + strconv.FormatInt(lastID, 32) //将文件名称转化为拼音
	//循环将每个章节的URL
	for _, data := range zhang_new_url {
		// fmt.Printf("获取的章节= %s \n", data)

		go Dowork(data, book_name, page, db, lastID)

		fmt.Printf("采集完成 url=%s ,%s\n", data, <-page)
		// return //暂时先停止

	}

	f, err := os.Create(filename + "_InsertSql.txt")
	if err != nil {
		fmt.Printf("os.Create ", err)
		return
	}

	_, err3 := f.WriteString(InsertSql)
	if err3 != nil {
		fmt.Printf("file.Write 出现问题 %s \n ", err3)
		return
	}
	// fmt.Printf("tmp = #%s# \n", tmp) //打印出来
	defer f.Close()

	// title, author, book_name, novel_status, description, image
	bookSql := "insert into n_book(title,img,author,`describe`,content,c_id,create_time,book_status) values('" + book_name + "','" + image + "','" + author + "','" + description + "','" + description + "',16,'" + time_at + "','" + novel_status + "');"

	ff, err := os.Create(filename + "_sql.txt")
	if err != nil {
		fmt.Printf("os.Create ", err)
		return
	}

	_, err_ := ff.WriteString(bookSql)
	if err_ != nil {
		fmt.Printf("file.Write 出现问题 %s \n ", err_)
		return
	}
	// fmt.Printf("tmp = #%s# \n", tmp) //打印出来
	defer ff.Close()
	fmt.Printf("book sql 语句是: %s \n", bookSql)

	//获取生成文件的信息大小、字数

	wenjian, err_f := os.Open(filename + ".txt")
	if err_f != nil {
		fmt.Printf("os.Open 错误 %s \n", err_f)
		return
	}

	defer wenjian.Close()

	file_buf := make([]byte, 1024*1024*50)
	n, err_file := wenjian.Read(file_buf)
	if n == 0 && err_file == io.EOF {
		fmt.Printf("wenjian.Read 错误 %s \n", err_file)
		return
	}

	filesize := strings.TrimSpace(tool.Getfilesize(filename + ".txt"))
	wordNum := tool.GetfontNum(string(file_buf[:n]))
	txt_url := filename + ".txt"
	//更新书中的txt大小、txt链接地址、字数
	stmt, err_book := db.Prepare("update n_book set word_number=?,book_size=?,txt_url=? where id=?")
	checkErr(err_book)

	_, errs := stmt.Exec(wordNum, filesize, txt_url, lastID)
	checkErr(errs)

	db.Close()
}
