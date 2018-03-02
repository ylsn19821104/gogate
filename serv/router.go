package serv

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/wanghongfei/gogate/utils"
	"gopkg.in/yaml.v2"
)

type Router struct {
	routePath	string
	routeMap	*sync.Map
}

type ServiceInfo struct {
	Id		string
	Path	string
}

/*
* 创建路由器
*
* PARAMS:
*	- path: 路由配置文件路径
*
*/
func NewRouter(path string) (*Router, error) {
	routeMap, err := loadRoute(path)
	if nil != err {
		return nil, err
	}

	return &Router{
		routeMap: routeMap,
		routePath: path,
	}, nil
}

/*
* 重新加载路由器
*/
func (r *Router) ReloadRoute() error {
	newRoute, err := loadRoute(r.routePath)
	if nil != err {
		return err
	}

	r.refreshRoute(newRoute)

	return nil
}

/*
* 将路由信息转换成string返回
*/
func (r *Router) ExtractRoute() string {
	var strBuf bytes.Buffer
	r.routeMap.Range(func(key, value interface{}) bool {
		strKey := key.(string)
		info := value.(*ServiceInfo)

		str := fmt.Sprintf("%s -> id:%s, path:%s\n", strKey, info.Id, info.Path)
		strBuf.WriteString(str)

		return true
	})

	return strBuf.String()
}

func (r *Router) refreshRoute(newRoute *sync.Map) {
	exclusiveKeys := utils.FindExclusiveKey(r.routeMap, newRoute)
	utils.DelKeys(r.routeMap, exclusiveKeys)
	utils.MergeSyncMap(newRoute, r.routeMap)
}

func loadRoute(path string) (*sync.Map, error) {
	// 打开配置文件
	routeFile, err := os.Open(path)
	if nil != err {
		return nil, err
	}
	defer routeFile.Close()

	// 读取
	buf, err := ioutil.ReadAll(routeFile)
	if nil != err {
		return nil, err
	}

	// 解析yml
	ymlMap := make(map[string]*ServiceInfo)
	err = yaml.UnmarshalStrict(buf, &ymlMap)
	if nil != err {
		return nil, err
	}


	// 构造 path->serviceId 映射
	var routeMap sync.Map
	for name, info := range ymlMap {
		// 验证
		err = validateServiceInfo(info)
		if nil != err {
			return nil, errors.New("invalid config for " + name + ":" + err.Error())
		}

		routeMap.Store(info.Path, info)
	}

	return &routeMap, nil
}

func validateServiceInfo(info *ServiceInfo) error {
	if nil == info {
		return errors.New("info is empty")
	}

	if "" == info.Id {
		return errors.New("id is empty")
	}

	if "" == info.Path {
		return errors.New("path is empty")
	}

	return nil
}