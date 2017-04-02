package main

import (
	"fmt"
	//"reflect"
	"runtime"
	"net/url"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var (
	mc *MySqlConfig
)

type Mobile struct {
	id     int
	mobile string
}

type MySqlConfig struct {
	Host    string
	MaxIdle int
	MaxOpen int
	User    string
	Pwd     string
	DB      string
	Port    int
	pool    *sql.DB
}

func (mc *MySqlConfig) Init() (err error) {
	// 构建 DSN 时尤其注意 loc 和 parseTime 正确设置东八区，允许解析时间字段
	url := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=%s&parseTime=true",
		mc.User,
		mc.Pwd,
		mc.Host,
		mc.Port,
		mc.DB,
		url.QueryEscape("Asia/Shanghai"),
	)

	// 全局实例只需调用一次
	mc.pool, err = sql.Open("mysql", url)
	if err != nil {
		return err
	}

	// 使用前 Ping，确保 DB 连接正常
	err = mc.pool.Ping()
	if err != nil {
		return err
	}

	// 设置最大连接数，一定要设置 MaxOpen
	mc.pool.SetMaxIdleConns(mc.MaxIdle)
	mc.pool.SetMaxOpenConns(mc.MaxOpen)
	return nil
}

func init() {
	mc = &MySqlConfig{
		Host:    "localhost",
		MaxIdle: 1000,
		MaxOpen: 2000,
		User:    "testuser",
		Pwd:     "testpwd",
		DB:      "mobiles",
		Port:    3306,
	}

	err := mc.Init()
	if err != nil {
		panic(err)
	}
}

func getAll(sql string, param ...interface{}) ([]*Mobile, error){
	var mobiles []*Mobile

	fmt.Println(param)
	//fmt.Println(reflect.TypeOf(param))

	err := mc.pool.Ping()
	if err != nil {
		return mobiles, err
	}

	stmt, err := mc.pool.Prepare(sql)
	if err != nil {
		return mobiles, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(param...)
	if err != nil {
		return mobiles, err
	}
	defer rows.Close()
	//fmt.Println(reflect.TypeOf(rows))

	for rows.Next() {
		mobile := &Mobile{}
		err = rows.Scan(&mobile.id, &mobile.mobile)
		if err != nil {
			continue
		}
		mobiles = append(mobiles, mobile)
	}

	return mobiles, nil
}

func getAll1(sql string) (map[interface{}]interface{}, error){
	mobiles := make(map[interface{}]interface{})
	var mobile Mobile

	err := mc.pool.Ping()
	if err != nil {
		return mobiles, err
	}

	rows, err := mc.pool.Query(sql)
	if err != nil {
		return mobiles, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&mobile.id, &mobile.mobile)
		if err != nil {
			continue
		}
		mobiles[mobile.id] = mobile
	}
	//sort.Sort(mobiles)

	return mobiles, nil
}

func main() {
	maxProcs := runtime.NumCPU() // 获取cpu个数
	runtime.GOMAXPROCS(maxProcs) //限制同时运行的goroutines数量

	fmt.Println("数据库初始化完成!")

	mobiles, err := getAll("SELECT id,mobile FROM t_mobile WHERE mobile LIKE ? LIMIT ?,?", "131%", 20, 20)
	if err != nil{
		panic(err)
	}

	//fmt.Printf("%v", mobiles)
	//fmt.Println(reflect.TypeOf(mobiles).Elem())
	//fmt.Println(reflect.ValueOf(mobiles))

	for _, mobile := range mobiles {
		fmt.Println((*mobile).id, (*mobile).mobile)
	}

	infos, err1 := getAll1("SELECT id,mobile FROM t_mobile ORDER BY id ASC LIMIT 50")
	if err1 != nil{
		panic(err1)
	}
	fmt.Println(infos)
}