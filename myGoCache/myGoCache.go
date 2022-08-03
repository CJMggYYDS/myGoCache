package myGoCache

import (
	"fmt"
	"log"
	"sync"
)

/*
	myGoCache.go
	框架入口,负责与用户交互,并且控制缓存值存取和获取的流程
*/

/*
	前置工作：
	设计回调Getter,如何从数据源获取数据交给用户自行解决
	在缓存不存在时,调用这个回调函数从数据源得到源数据
*/
// Getter 定义接口Getter
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 定义函数类型GetterFunc
type GetterFunc func(key string) ([]byte, error)

/*
	让函数类型GetterFunc实现Getter接口的Get方法,使其成为接口型函数
	方便使用者在调用时既能传入函数作为参数，也能传入实现了该接口的结构体作为参数
	*定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
*/
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

/*
    核心结构Group负责的工作流程:
                               是
	接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
                   |  否                         是
                   |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
                                |  否
                                |-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

*/

// Group 定义核心数据结构Group,负责与用户交互,并且控制缓存值的存储和获取
type Group struct {
	name      string //一个Group相当于一个缓存的命名空间,每个Group拥有一个唯一的名称
	getter    Getter //缓存未命中时获取源数据的回调(callback)
	mainCache cache  //支持并发读写的缓存
}

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup Group的构造函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mutex.Lock()
	defer mutex.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup 用于获取特定名称的Group
func GetGroup(name string) *Group {
	mutex.RLock() //这里因为只读，故使用了只读锁
	g := groups[name]
	mutex.RUnlock()
	return g
}

// Get -- Group的核心方法,接收key来查找缓存值。
//   先从mainCache中查找, 如果存在直接返回值;
//   如果缓存不存在, 则调用load方法, load调用getLocally, getLocally调用用户回调函数g.getter.Get()从本地数据源获取源数据并添加到mainCache中。
// 目前暂时不考虑分布式情况
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[myGoCache] hit")
		return v, nil
	}
	return g.load(key)
}

// 目前只考虑从本地获取源数据
func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
