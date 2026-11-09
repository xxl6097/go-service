### ✨ 新特性

* 将可执行程序以服务形式安装，支持跨平台linux、macos平台；

---
### 🐛 Bug 修复（签名长度越界，会产出无法启动的二进制）

**现象**：安装/升级后程序一启动就 panic，栈顶是

```
fatal error: index out of range
runtime.getHugePageSize()  runtime/os_linux.go:331
```

**根因**：`ukey` 用「原地替换占位符」的方式把配置写进可执行文件，占位区是嵌入的
一段 `+`。`GenConfig` 的容量校验只算了 `body`，漏算 `rawKey(39) + 4字节长度头`，
且数据超出占位区时既不报错也不补齐；`GenerateBin` 又直接按新长度覆盖
→ 文件被"撑大"，但 ELF 的段表偏移不会更新，文件尾部整体位移，
运行时全局变量（`sysTHPSizePath`）读错位、len 变 0，`&sysTHPSizePath[0]` 越界 panic。
（实测：占位区 1024，写入 1041 字节，文件 +17 字节。）

**修复**：

* `GenConfig` 容量校验改为 `rawKey + 4 + body`，超界直接返回错误，绝不静默撑大；
* `GenerateBin` 增加等长硬校验，签名数据与占位区长度不一致时直接终止，并校验产物长度与源文件一致；
* `GetCfgBufferFromFile` 改为按文件内的长度头解析签名区，不再依赖 `len(buffer)`，
  占位区跨版本扩容（1024 → 8192）也能正确提取；
* 新增 `NormalizeBuffer`，把旧版本提取出的较短签名数据补齐到新占位区长度；
* `SignFileByBuffer` 校验签名数据必须以 rawKey 开头，避免写入"长度对、内容错"的坏程序；
* `DecKey` 增加长度校验，避免短数据 panic；
* 内置占位区 `assets/buffer/buffer` 由 1024 扩到 8192。

---
### 🚀 github加速

```
[
  "https://ghfast.top/",
  "https://gh-proxy.com/",
  "https://ghproxy.1888866.xyz/",
  "https://github.moeyy.xyz/"
]
```