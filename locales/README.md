### 国际化翻译配置
##### 服务需要一个简单的国际化支持
* 文件名必须符合 BCP 47 tag 标准, 可打印 language.Chinese, language.English 等 tag 的字符串获得
* 理论支持 toml, json, yaml 等, 只要指定扩展包提供 Unmarshal func(data []byte, v interface{}) error 方法
* 使用 text/template 解析字符串模板, 所以可以传入map,结构体等填充字符串
* 支持自定义任一复数形式模板

#### example: 
```toml
# locales/admin/en.toml
[Hello]
description = "这是一个示例的模板"
one = "{{.name}} has {{.count}} cat."
zero = "{{.name}} has no cat"
few = "{{.name}} has few cat"
other = "{{.name}} has {{.count}} cats." # other 必填
```

```go
package main

import (
    "fmt"
    "github.com/BurntSushi/toml"
    "golang.org/x/text/language"
    "starter/pkg/i18n"  
)

func main(){ 
    // 实例化一个新的Bundle并加载locales/admin 目录下的所有模板文件,以文件名作为tag
    bundle := i18n.NewBundle(language.English).LoadFiles("./locales/admin", toml.Unmarshal) 
    fmt.Println(bundle.NewPrinter(language.English).Translate("Hello", i18n.Data{"name": "medivh", "count": 1}, 1))
    // output:  medivh has 1 cat.
}
```