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

```bash
   go get https://github.com/iEvan-lhr/exciting-tool
```    

## full-featured String

Use the following code to construct a full -featured String
   
```bash
    tools.Make(str)
```    
Common Functions

✔The following functions support the use of String, full -featured String, [] byte as the parameters

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
