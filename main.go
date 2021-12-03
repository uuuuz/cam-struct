package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var reg = regexp.MustCompile(`.*cam/back/.*`) // cam/back/

type (
	Node struct {
		Name     string  `json:"name"`
		Children []*Node `json:"children"`
	}

	Depends []string
)

func (s Depends) Len() int           { return len(s) }
func (s Depends) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Depends) Less(i, j int) bool { return s[i] < s[j] }

func main() {
	// 生成树状结构
	http.HandleFunc("/dep", getTree)

	// 需要安装 brew install graphviz
	http.HandleFunc("/all", getRelation)

	http.HandleFunc("/simple", getSimpleRelation)

	if err := http.ListenAndServe(":3001", nil); err != nil {
		fmt.Println(err.Error())
	}
}

func getTree(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*")                // 跨域
	writer.Header().Set("Content-type", "application/json; charset=utf-8") // 返回json

	pkgs, data, err := getDepends()
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 树状结构
	root := &Node{Name: "back", Children: make([]*Node, len(pkgs))}
	for i := range pkgs {
		v := data[pkgs[i]]
		child := &Node{Name: pkgs[i], Children: []*Node{}}
		for n := range v {
			leaf := &Node{Name: v[n]}
			child.Children = append(child.Children, leaf)
		}
		root.Children[i] = child
	}
	res, err := json.Marshal(root)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	//fmt.Println(string(res))
	// 关系图
	// res, _ := getResultData(pkgs, data)
	if _, err := writer.Write(res); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getRelation(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file

	target := request.FormValue("name")

	_, data, err := getDepends()
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 生成文件
	buf := &bytes.Buffer{}
	for k := range data {
		if k != target {
			continue
		}
		buf.WriteString("digraph G {\n")
		combineByFloor(buf, k, map[string]struct{}{}, map[string]struct{}{}, data)
		buf.WriteString("}")
		break
	}
	if _, err := writer.Write(buf.Bytes()); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getSimpleRelation(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file
	target := request.FormValue("name")

	_, data, err := getDepends()
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 生成文件
	buf := &bytes.Buffer{}
	for k := range data {
		if k != target {
			continue
		}
		buf.WriteString("digraph G {\n")
		simpleCombine(buf, k, data)
		buf.WriteString("}")
		break
	}
	if _, err := writer.Write(buf.Bytes()); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

type node struct {
	name   string
	father map[string]*node
	child  map[string]*node
}

func simpleCombine(buf *bytes.Buffer, rootK string, source map[string][]string) {
	root := &node{
		name:  rootK,
		child: make(map[string]*node),
	}
	nodePool := map[string]*node{"user": root}
	// 找出所有依赖关系
	var fathers = []*node{root}
	for len(fathers) > 0 {
		fathers = getDepend(fathers, source, nodePool)
	}
	// 去掉多余的可达路线
	var routes = []*node{root}
	for len(routes) > 0 {
		routes = filterDepend(routes)
	}
	// 绘制依赖关系 - 去掉重复的依赖关系
	relation := make(map[string]struct{})
	var next = []*node{root}
	for {
		next = draw(buf, next, relation)
		if len(next) == 0 {
			break
		}
	}
}

func filterDepend(roots []*node) []*node {
	if roots == nil {
		return nil
	}
	var next []*node
	for i := range roots {
		root := roots[i]
		for k, v := range root.child {
			routes := make([]*node, 0, len(root.child)-1)
			for kk := range root.child {
				if kk != k {
					routes = append(routes, root.child[kk])
				}
			}
			if reachable(v, routes) {
				delete(root.child, k)
			} else {
				next = append(next, v)
			}
		}
	}
	return next
}

func reachable(target *node, routes []*node) bool {
	if target == nil || len(routes) == 0 {
		return false
	}
	for i := range routes {
		if canArrive(target.name, routes[i]) {
			return true
		}
	}
	return false
}

func draw(buf *bytes.Buffer, nodes []*node, relation map[string]struct{}) []*node {
	var next []*node
	for i := range nodes {
		for _, v := range nodes[i].child {
			str := fmt.Sprintf("%s -> %s\n", nodes[i].name, v.name)
			if _, exist := relation[str]; exist {
				continue
			}
			buf.WriteString(str)
			relation[str] = struct{}{}
			next = append(next, v)
		}
	}
	return next
}

func getDepend(fathers []*node, source map[string][]string, nodePool map[string]*node) []*node {
	if len(fathers) == 0 {
		return nil
	}
	var next []*node
	for i := range fathers {
		father, children := fathers[i], source[fathers[i].name]
		for ii := range children {
			child := nodePool[children[ii]]
			if child == nil {
				nodePool[children[ii]] = &node{
					name:   children[ii],
					father: map[string]*node{},
					child:  map[string]*node{},
				}
				child = nodePool[children[ii]]
			}
			// 是否构成了环
			if canArrive(father.name, child) {
				// 若为环，则不构建关系
				continue
			}
			// 建立关系
			child.father[father.name] = father
			father.child[child.name] = child
			if child.name == "startup" {
				continue
			}
			next = append(next, child)
		}
	}
	return next
}

func canArrive(father string, child *node) bool {
	if child.name == father || child == nil {
		return true
	}
	for _, v := range child.child {
		if canArrive(father, v) {
			return true
		}
	}
	return false
}

func combineByFloor(buf *bytes.Buffer, father string, notRef, relation map[string]struct{}, source map[string][]string) {
	// 逐层生成
	var floor = []string{father}
	for {
		var next []string
		for i := range floor {
			next = append(next, combineChild(buf, floor[i], source[floor[i]], notRef, relation)...)
		}
		if len(next) == 0 {
			break
		}
		floor = next
	}
}

func combineChild(buf *bytes.Buffer, father string, child []string, notRef, relation map[string]struct{}) []string {
	if len(child) == 0 {
		return nil
	}
	var next []string
	for i := range child {
		// 排除不可引用的
		if father == child[i] {
			continue
		}
		if _, exist := notRef[child[i]]; exist {
			continue
		}
		str := fmt.Sprintf("%s -> %s\n", father, child[i])
		if _, exist := relation[str]; exist {
			continue
		}
		buf.WriteString(str)
		relation[str] = struct{}{}

		if child[i] == "startup" {
			continue
		}
		next = append(next, child[i])
	}
	return next
}

func getDepends() ([]string, map[string][]string, error) {
	// 遍历cam项目，拉取所有的包和依赖
	path := "/Users/wxm/test/cam/back/"
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, nil, err
	}
	var pkgs []string // 是有序的
	depends := make(map[string][]string)
	for i := range files {
		if !files[i].IsDir() {
			continue
		}
		dep, err := readModule(path + files[i].Name())
		if err != nil {
			return nil, nil, err
		}
		depends[files[i].Name()] = dep
		pkgs = append(pkgs, files[i].Name())
	}
	return pkgs, depends, nil
}

func readModule(path string) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var depends []string
	for i := range files {
		fileName := path + "/" + files[i].Name()
		var dep []string
		if files[i].IsDir() {
			dep, err = readModule(fileName)
			if err != nil {
				return nil, err
			}
		} else {
			data, err := ioutil.ReadFile(fileName)
			if err != nil {
				return nil, err
			}
			lines := strings.Split(string(data), "\n")
			for n := range lines {
				if strings.HasPrefix(lines[n], ")") || strings.HasPrefix(lines[n], "func") {
					break
				}
				line := reg.FindString(lines[n])
				if len(line) > 0 {
					line = strings.Split(line, "cam/back/")[1]
					line = strings.Trim(line, `"`)

					dep = append(dep, line)
				}
			}
		}
		if len(dep) > 0 {
			depends = append(depends, dep...)
		}
	}
	// 去重
	depMap := make(map[string]struct{})
	for i := range depends {
		// depMap[depends[i]] = struct{}{}
		depMap[strings.Split(depends[i], "/")[0]] = struct{}{}
	}
	res := make([]string, 0, len(depMap))
	for k := range depMap {
		res = append(res, k)
	}
	// 排序 todo
	sort.Sort(Depends(res))
	return res, nil
}
