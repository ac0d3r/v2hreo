<div align="center" >
    <h1>v2hreo</h1>
    <img src="https://user-images.githubusercontent.com/26270009/125184385-cfb56f00-e24f-11eb-8983-cf97c3189e0f.png" width="100" alt="v2hreo" />
</div>

Swift 联动 CGO 开发的 V2ray MacOS 菜单栏应用。

---

支持的功能:
- **默认配置:** `socks://127.0.0.1:1080` 目前还不支持修改
- **订阅地址:** `vmess://*` 或者 `http://*`(返回也的是`vmess://`)
- **服务器选择:** 
    - Load: 从订阅地址获取服务器列表
    - Ping: 获取连接服务器延迟
    - 下拉选择代理的服务器

---

## 预览

![](https://user-images.githubusercontent.com/26270009/125184897-0097a300-e254-11eb-973b-8970549c1a8f.png)

## Tips  
**自己创建项目时需要注意：**
1. 编译 go 代码：`CGO_ENABLED=1 go build --buildmode=c-archive  -o libdemo.a demo.go`
2. 在 Swift 项目`$(SRCROOT)`目录下创建 `module.modulemap` 文件：
    ```
    module Demo {
        header "libdemo.h"
        link "demo"
        export *
    }
   ```
    具体请查看[文档](https://clang.llvm.org/docs/Modules.html)
3. 为 Swift 项目设置 modulemap：在 Xcode 中将 `LIBRARY_SEARCH_PATHS` 和 `SWIFT_INCLUDE_PATHS` 的值修改为：`$(SRCROOT)`。

## Thx
- https://youngdynasty.net/posts/writing-mac-apps-in-go/
- https://juejin.cn/post/6844904101877121037
