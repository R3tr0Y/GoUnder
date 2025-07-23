# GoUnder

GoUnder 是一款由go语言开发的前期渗透信息搜集工具，目前主要由两个模块组成：fingerprint 和 cdn 模块。

1. fingeprint

   用法：

   ```
   GoUnder fingerprint -u <URL> [flag]
   ```

   该模块用于自动分析网站的指纹，目前有两个引擎可供选择：wappalyzergo、whatcms。wappalyzergo是利用开源的特征库在本地识别，主要通过请求头、cookie、网站前端代码的内容对指纹特征库正则匹配；whatcms引擎，是利用whatcms提供的api在线查询网站的技术指纹，需联网使用，并且要自己配置API key。

2. cdn

   用法

   ```
   GoUnder cdn -u <URL> [flag]
   ```

   该模块用于自动绕过存在CDN防护的网站，寻找网站的真实IP地址，主要通过FOFA网络空间搜索引擎的语法技巧自动化分析，需要配置FOFA账号的api key，并且联网使用。拥有三种绕过方案，host、title、icon。

3. webui
   用法

   ```
   GoUnder webui [-u host -p port]
   ```
   该命令用于启动webui页面，为用户提供可视化操作