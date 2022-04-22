package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

var (
	reg          = regexp.MustCompile(`.*cam/back/.*`) // cam/back/
	rootPath     = "/Users/wxm/go/src/cam/back/"
	rootNodeName = "root_root_root"
)

// hi how are you

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
	// 默认包名与文件夹名一致 //
	// 需要安装 brew install graphviz
	http.HandleFunc("/all-all", getRelation)
	http.HandleFunc("/all-simple", getRelationSimple)
	http.HandleFunc("/single-simple", getSingleRelationSimple)

	// 最外层的依赖关系
	http.HandleFunc("/all", getAll)
	http.HandleFunc("/simple", getAllSimple)

	if err := http.ListenAndServe(":3001", nil); err != nil {
		fmt.Println(err.Error())
	}
}

func getAll(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file
	// 扫描全部依赖
	_, data, err := getDepends(rootPath)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 找出所有根结点（不被依赖的所有包）
	var rootMap []string
	for k := range data {
		isDepended := false
	OVER:
		for _, vv := range data {
			for i := range vv {
				if vv[i] == k {
					isDepended = true
					break OVER
				}
			}
		}
		if !isDepended {
			rootMap = append(rootMap, k)
		}
	}
	// 当根结点不止1个时，创建一个结点（虚构的，如果可以则不进行创建），依赖这些根节点
	if len(rootMap) > 0 {
		data[rootNodeName] = rootMap
	}
	// 遍历依赖结构
	buf := &bytes.Buffer{}
	buf.WriteString("digraph G {\n")
	combineByFloor(buf, rootNodeName, map[string]struct{}{}, data)
	buf.WriteString("}")

	if _, err := writer.Write(buf.Bytes()); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getAllSimple(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file

	_, data, err := getDepends(rootPath)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 找出所有根结点（不被依赖的所有包）
	var rootMap []string
	for k := range data {
		isDepended := false
	OVER:
		for _, vv := range data {
			for i := range vv {
				if vv[i] == k {
					isDepended = true
					break OVER
				}
			}
		}
		if !isDepended {
			rootMap = append(rootMap, k)
		}
	}
	// 当根结点不止1个时，创建一个结点（虚构的，如果可以则不进行创建），依赖这些根节点
	if len(rootMap) > 0 {
		data["root_root_root"] = rootMap
	}

	// 生成文件
	buf := &bytes.Buffer{}
	buf.WriteString("digraph G {\n")
	simpleCombine(buf, "root_root_root", data)
	buf.WriteString("}")
	if _, err := writer.Write(buf.Bytes()); err != nil {
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
	// 扫描全部依赖
	data, err := getRelationDepends(rootPath)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 找出所有根结点（不被依赖的所有包）
	var rootMap []string
	for k := range data {
		isDepended := false
	OVER:
		for _, vv := range data {
			for i := range vv {
				if vv[i] == k {
					isDepended = true
					break OVER
				}
			}
		}
		if !isDepended {
			rootMap = append(rootMap, k)
		}
	}
	// 当根结点不止1个时，创建一个结点（虚构的，如果可以则不进行创建），依赖这些根节点
	if len(rootMap) > 0 {
		data[rootNodeName] = rootMap
	}
	// 遍历依赖结构
	buf := &bytes.Buffer{}
	buf.WriteString("digraph G {\n")
	combineByFloor(buf, rootNodeName, map[string]struct{}{}, data)
	buf.WriteString("}")

	if _, err := writer.Write(buf.Bytes()); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getRelationSimple(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file

	data, err := getRelationDepends(rootPath)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 找出所有根结点（不被依赖的所有包）
	var rootMap []string
	for k := range data {
		isDepended := false
	OVER:
		for _, vv := range data {
			for i := range vv {
				if vv[i] == k {
					isDepended = true
					break OVER
				}
			}
		}
		if !isDepended {
			rootMap = append(rootMap, k)
		}
	}
	// 当根结点不止1个时，创建一个结点（虚构的，如果可以则不进行创建），依赖这些根节点
	if len(rootMap) > 0 {
		data["root_root_root"] = rootMap
	}

	// 生成文件
	buf := &bytes.Buffer{}
	buf.WriteString("digraph G {\n")
	simpleCombine(buf, "root_root_root", data)
	buf.WriteString("}")
	if _, err := writer.Write(buf.Bytes()); err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

func getSingleRelationSimple(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Access-Control-Allow-Origin", "*") // 跨域
	writer.Header().Add("Content-Disposition", "attachment; filename=demo")
	writer.Header().Set("Content-type", "text/plain") // 返回 file
	target := request.FormValue("name")

	data, err := getRelationDepends(rootPath)
	if err != nil {
		fmt.Println(err.Error())
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	for k := range data {
		if strings.Compare(k, target) == 0 {
			// 生成文件
			buf := &bytes.Buffer{}
			buf.WriteString("digraph G {\n")
			simpleCombine(buf, k, data)
			buf.WriteString("}")
			if _, err := writer.Write(buf.Bytes()); err != nil {
				fmt.Println(err.Error())
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
			break
		}
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
	nodePool := map[string]*node{rootK: root}
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
	relationMap := make(map[string]struct{})
	label := make(map[string]struct{})
	var next = []*node{root}
	var relationSlice []string
	for {
		var relation []string
		next, relation = draw(next, relationMap, label)
		if len(next) == 0 {
			break
		}
		if len(relation) > 0 {
			relationSlice = append(relationSlice, relation...)
		}
	}
	for k := range label {
		if strings.HasPrefix(k, rootNodeName) {
			continue
		}
		buf.WriteString(k)
	}
	for i := range relationSlice {
		if strings.HasPrefix(relationSlice[i], rootNodeName) {
			continue
		}
		buf.WriteString(relationSlice[i])
	}
}

func filterDepend(roots []*node) []*node {
	var unique = make(map[string]struct{})
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
				if _, exist := unique[v.name]; exist {
					continue
				} else {
					unique[v.name] = struct{}{}
				}
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

func draw(nodes []*node, relation map[string]struct{}, labels map[string]struct{}) ([]*node, []string) {
	var next []*node
	var relationOrderly []string
	for i := range nodes {
		name, label := deal(nodes[i].name)
		if strings.Compare(name, label) != 0 {
			labels[fmt.Sprintf("%s[label=\"%s\"]\n", name, label)] = struct{}{}
		}
		for _, v := range nodes[i].child {
			n1, l1 := deal(v.name)
			if strings.Compare(n1, l1) != 0 {
				labels[fmt.Sprintf("%s[label=\"%s\"]\n", n1, l1)] = struct{}{}
			}
			str := fmt.Sprintf("%s -> %s\n", name, n1)
			if _, exist := relation[str]; exist {
				continue
			}
			// buf.WriteString(str)
			relationOrderly = append(relationOrderly, str)
			relation[str] = struct{}{}
			next = append(next, v)
		}
	}
	return next, relationOrderly
}

func deal(name string) (string, string) {
	label := name
	if strings.Contains(name, "/") {
		name = strings.Replace(name, "/", "___", -1)
		name = strings.Replace(name, "-", "__", -1)
	}
	return name, label
}

func getDepend(fathers []*node, source map[string][]string, nodePool map[string]*node) []*node {
	if len(fathers) == 0 {
		return nil
	}
	var unique = make(map[string]struct{})
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
			// if canArrive(father.name, child) {
			//	continue
			// }
			// 建立关系
			child.father[father.name] = father
			father.child[child.name] = child
			//if child.name == "startup" {
			//	continue
			//}
			// 去重
			if _, exist := unique[child.name]; exist {
				continue
			} else {
				unique[child.name] = struct{}{}
			}
			next = append(next, child)
		}
	}
	return next
}

func canArrive(father string, child *node) bool {
	if child.name == father {
		return true
	}
	var next = child.child
	for len(next) > 0 {
		if _, exist := next[father]; exist {
			return true
		} else {
			nextMap := make(map[string]*node)
			for _, v := range next {
				for kk, vv := range v.child {
					nextMap[kk] = vv
				}
			}
			next = nextMap
		}
	}
	//for _, v := range child.child {
	//	if canArrive(father, v) {
	//		return true
	//	}
	//}
	return false
}

func combineByFloor(buf *bytes.Buffer, father string, relation map[string]struct{}, source map[string][]string) {
	// 逐层生成
	var floor = []string{father}
	for {
		var next []string
		for i := range floor {
			next = append(next, combineChild(buf, floor[i], source[floor[i]], relation)...)
		}
		if len(next) == 0 {
			break
		}
		floor = next
	}
}

func combineChild(buf *bytes.Buffer, father string, child []string, relation map[string]struct{}) []string {
	if len(child) == 0 {
		return nil
	}
	var next []string
	for i := range child {
		str := fmt.Sprintf("%s -> %s\n", father, child[i])
		if _, exist := relation[str]; exist {
			continue
		}
		if strings.Compare(father, rootNodeName) != 0 {
			buf.WriteString(str)
		}
		relation[str] = struct{}{}
		//if child[i] == "startup" {
		//	continue
		//}
		next = append(next, child[i])
	}
	return next
}

func getDepends(path string) ([]string, map[string][]string, error) {
	// 遍历cam项目，拉取所有的包和依赖
	// path := "/Users/wxm/test/cam/back/" // 做成参数 todo
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
			continue // 暂且过滤内部的文件夹
			//dep, err = readModule(fileName)
			//if err != nil {
			//	return nil, err
			//}
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
		// depMap[depends[i]] = struct{}{} // 取全包含的结构，考虑需要替换的敏感字符
		if !strings.Contains(depends[i], "/") { // 仅取最外层结构
			depMap[depends[i]] = struct{}{}
		}
	}
	res := make([]string, 0, len(depMap))
	for k := range depMap {
		res = append(res, k)
	}
	// 按照名字排序
	sort.Sort(Depends(res))
	return res, nil
}

// ***********************
// 所有层面的结构
// ***********************
func getRelationDepends(path string) (map[string][]string, error) {
	// 遍历cam项目，拉取所有的包和依赖
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	depends := make(map[string][]string)
	for i := range files {
		if !files[i].IsDir() {
			continue
		}
		dep, err := readRelationModule(path+files[i].Name(), files[i].Name())
		if err != nil {
			return nil, err
		}
		for k, v := range dep {
			if _, exist := depends[k]; exist {
				depends[k] = append(depends[k], v...)
			} else {
				depends[k] = v
			}
		}
	}

	return depends, nil
}

func readRelationModule(path, pkg string) (map[string][]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	allDepends := make(map[string][]string) // 该路径下所有依赖关系
	var depends []string                    // 当前包的依赖关系
	for i := range files {
		fileName := path + "/" + files[i].Name()
		if files[i].IsDir() {
			deps, err := readRelationModule(fileName, pkg+"/"+files[i].Name())
			if err != nil {
				return nil, err
			}
			for k, v := range deps {
				allDepends[k] = v
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
					depends = append(depends, line)
				}
			}
		}
	}
	// 去重
	depMap := make(map[string]struct{})
	for i := range depends {
		depMap[depends[i]] = struct{}{} // 取全包含的结构，考虑需要替换的敏感字符
	}
	res := make([]string, 0, len(depMap))
	for k := range depMap {
		res = append(res, k)
	}
	// 按照名字排序
	sort.Sort(Depends(res))
	allDepends[pkg] = res
	return allDepends, nil
}
