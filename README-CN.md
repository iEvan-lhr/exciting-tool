# exciting-tool

[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://opensource.org/licenses/Apache-2.0)

[English](./README.md) | [简体中文](./README-CN.md)

exciting-tool 是一个理想化的全能的 Go 的工具包，涉及的内容包括:字符串的处理（拼接、转换等等）、结构体的日志遍历（示例）、原生sql的自动拼接与处理、更方便的map（开发中）、json解析、http请求发送等等。


## 备注信息

这是一个完全开源的项目。

作者并不能保证能及时更新内容，但会尽力去完善文档来帮助使用者来使用。

也欢迎大家帮助我 :heart::heart::heart:

## 重要提示  ⚠⚠⚠⚠⚠


☠ 需要GO版本1.18或更多 ☠

☀ 您可以在低版本中使用Lowversion分支的代码 ☀

某些功能可能不安全，并且某些功能仅在GO原生代码中修改。因此，请注意实际使用中某些功能的使用。在可能的问题上，我会尽可能地标记。

## 如何使用

```bash
   go get github.com/iEvan-lhr/exciting-tool
```    
## 全功能 String  

### 相同的功能

✔以下功能支持使用String，全功能String，[]byte，部分支持rune

使用这段代码来构造全功能String

```bash
    tools.Make(str)
```  


```plain

Function:
  Index(str any)           The next bid search, while supporting the Rune type retrieval
  Append(join any)         Add content to the string to support adding 
                           all basic types and extension basic types 
                           (including int, float, BOOL, int32, int16, string, str, byte, [] byte ...). 
                           Can be added (PS: pointer is passed in)
  Make(obj any)            If the structure is used to construct and the structure 
                           does not implement the String () method,
                           the full attribute printing will be performed. 
                           The example is as follows:
                           ----------User----------
                           Id:23132
                           Username:foo
                           Password:bar
                           Identity:324213
                           QrCode:982j32
                           DenKey:ansssss
                           TalkingKey:qwesad
                           ----------END----------
  FirstUpper()
  FirstLower()
  Check(str any)
  RemoveLastStr(lens)
  RemoveIndexStr(lens)
  Spilt(str any)
  CheckIsNull()

```

## Error treatment

```plain

Function:
  ReturnValueByTwo()       The return value after the automatic processing, 
                           if the error is not empty, will panic(err)
  ReturnValue()            The return value after the automatic processing, 
                           if the error is not empty, will log(err)
  ExecGoFunc()             The error task that can be automatically defined in the asynchronous 
                           execution method internally is the asynchronous
                           execution of the error task that may occur
  ExecError()
  PanicError()
  logError()

```
