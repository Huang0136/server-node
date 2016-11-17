package service

import (
	"bytes"
	"constants"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"logs"
	"net/http"
	"net/rpc"
	"os"
	"reflect"
	"service/impl"
	"validation"
)

// 服务节点
type ServiceNode int

// 请求参数
type Req struct {
	Params map[string]interface{}
}

// 返回数据
type Resp struct {
	Params map[string]interface{}
}

// 服务接口结构体
type Service struct {
	ServiceId   string     `json:"service_id"`
	ServiceName string     `json:"service_name"`
	URL         string     `json:"url"`
	Regexp      bool       `json:"regexp"`
	MethodType  string     `json:"method_type"`
	Method      string     `json:"method"`
	Remark      string     `json:"remark"`
	InParams    []InParam  `json:"in_params"`
	OutParams   []OutParam `json:"out_params"`
}

// 入参结构体
type InParam struct {
	ParamCode string `json:"param_code"`
	ParamName string `json:"param_name"`
	ParamType string `json:"param_type"`
	Require   bool   `json:"requrie"`
	Length    int    `json:"length"`
	Remark    string `json:"remark"`
}

// 出参结构体
type OutParam struct {
	ParamCode string `json:"param_code"`
	ParamName string `json:"param_name"`
	ParamType string `json:"param_type"`
	Remark    string `json:"remark"`
}

// 服务接口列表
var Services []Service = make([]Service, 0)

//
func init() {
	// 读取服务接口配置文件
	readServiceConfig()

	// 打印服务接口
	//	PrintlnServices()
}

// 打印服务接口列表
func PrintlnServices() {
	for _, s := range Services {
		fmt.Println(ServiceToStr(s))
	}
}

// 将服务转换成字符串
func ServiceToStr(s Service) string {
	b := bytes.Buffer{}

	b.WriteString("接口:" + s.ServiceName + ",")
	b.WriteString("方法:" + s.Method)

	return b.String()
}

// 统一的rpc调用处理方法
func (serverNode *ServiceNode) RpcCallHandler(req Req, resp *Resp) error {
	methodName, _ := req.Params["METHOD_NAME"].(string) // 方法名称

	// 业务服务实现
	si := new(impl.ServiceImpl)
	si.InParams = req.Params
	fmt.Println("节点请求参数:", si.InParams)

	// 接口入参校验
	err := validation.CheckInterfaceInParams(methodName)
	if err != nil {
		return err
	}

	// 封装入参
	var inReflectValues []reflect.Value = make([]reflect.Value, 0)

	// 反射调用
	funcTemp := reflect.ValueOf(si).MethodByName(methodName)
	rfValues := funcTemp.Call(inReflectValues)

	fmt.Println("业务执行结果:", rfValues)

	// 返回结果
	resp.Params = make(map[string]interface{})
	resp.Params["result"] = rfValues[0].String()
	return nil
}

// 注册rpc
func RegisterRpc() {
	logs.MyDebugLog.Println("register rpc now...")

	serviceNode := new(ServiceNode)
	err := rpc.Register(serviceNode)
	logs.MyErrorLog.CheckFatallnError("", err)

	rpc.HandleHTTP()
	err = http.ListenAndServe(":"+constants.Configs["serverNode.port"], nil)
	logs.MyErrorLog.CheckFatallnError("", err)
	logs.MyDebugLog.Println("register rpc success...")
}

// 读取服务接口配置文件
func readServiceConfig() {
	serviceConfig, err := os.Open("./config/service.json")
	logs.MyDebugLog.CheckFatallnError("read service config file error:", err)

	sByte, err := ioutil.ReadAll(serviceConfig)
	logs.MyDebugLog.CheckFatallnError("read service config to byte error:", err)

	err = json.Unmarshal(sByte, &Services)
	logs.MyDebugLog.CheckFatallnError("convert byte to json error:", err)
}
