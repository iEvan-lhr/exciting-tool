# exciting-tool

[![Hex.pm](https://img.shields.io/hexpm/l/plug.svg)](https://opensource.org/licenses/Apache-2.0)

<h2>如何使用</h2>
<hr/>
<h3>String</h3>
<p>
    使用下面的方法来构造一个String
   
```bash
    tool.Strings(str)
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