# exciting-tool

[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://opensource.org/licenses/Apache-2.0)

[English](./README.md) | [简体中文](./README-CN.md)

Excing-Tool is an ideal omnidirectional GO toolkit. The content involved includes: string processing (stitching, conversion, etc.), the log traversal of the structure (example), the automatic stitching and processing of the native SQL, and the more Convenient Map (under development), JSON parsing, HTTP request sending, and so on.

## Remark information

The author does not guarantee to update the content in time, but will try his best to improve the document to help users use it.

Welcome everyone to help me  :heart::heart::heart:

## important hint ⚠⚠⚠⚠⚠

☠ Need GO version 1.18 or more ☠ 

☀ You can use the code of the Lowversion branch in the low version ☀

Some functions may be unsafe, and some functions are only modified based on the GO native bag. Therefore, please pay attention to the use of some functions in actual use. I will mark as much as possible where possible problems may occur.

## HOW TO USE

Use the following code to construct a full -featured String
   
```bash
    tools.Make(str)
```    
Common Functions

✔The following functions support the use of String, full -featured String, [] byte as the parameters

```plain
You should run "cf config" to configure your handle, password and code
templates at first.

If you want to compete, the best command is "cf race".

Function:
  Index(str any)           The next bid search, while supporting the Rune type retrieval
  Append(join any)         Add content to the string to support adding all basic types and extension basic types 
                           (including int, float, BOOL, int32, int16, string, str, byte, [] byte ...). 
                           Can be added (PS: pointer is passed in)
  Make(obj any)            If the structure is used to construct and the structure 
                           does not implement the String () method, the full attribute printing will be performed. 
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

```

支持的方法：<br/>
<a>Index(str)</a>  下标检索 str可以为string、byte、bytes、rune<br/>
<a>FirstUpper()</a>/<a>FirstLower()</a> 首字母大小写<br/>
<a>Check(str)</a> 比较 支持传入数据为string、bytes、rune<br/>
<a>Append(join)</a> 向字符串中添加内容 支持添加所有基本类型及扩展基本类型(包括int,float,bool,int32,int16,string,str,byte,[]byte...) 若结构体实现了String()方法也可以添加(ps:指针传入)<br/>
<a>RemoveLastStr(lens)</a>/<a>RemoveLastStrByRune()</a> 移除尾部的元素 <br/>
<a>RemoveIndexStr(lens)</a>/<a>RemoveIndexStrByRune()</a> 移除头部的元素 <br/>
<a>Spilt(str)</a> 按照str截取字符串 支持传入数据为string、bytes<br/>
<a>CheckIsNull()</a> 检查字符串是否为空 只包含' '与'\t'与'\n'都会被视为不合法的值<br/>

#### **......**

</p>

<hr/>
<h3>错误处理</h3>
<p>
支持的方法：<br/>
<a>ReturnValueByTwo</a>  
<a>ReturnValue</a> 传入返回值为两个的方法 返回首个元素 若错误不为空则会log(err)<br/>
<a>PanicError</a> 传入结束方法  支持多方法传入 例如 file.close() res.close() 若错误不为空则会panic(err)<br/>
<a>ExecError</a> 传入方法 获取返回值错误 若错误不为空则会panic(err)<br/>
<a>logError</a> 传入方法 获取返回值错误 若错误不为空则会log(err)<br/>
<a>ExecGoFunc</a> 传入异步执行方法 内部会自动defer捕捉方法可能出现的错误 任务为异步执行<br/>
</p>
