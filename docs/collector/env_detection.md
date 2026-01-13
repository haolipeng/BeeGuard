// 可疑的环境变量名
// 这些环境变量可能被攻击者利用进行权限提升、代码注入、数据泄露等攻击

1、LD_PRELOAD
作用：动态链接库预加载，可被用于劫持系统调用，实现权限提升或隐藏恶意行为
攻击场景: 设置 LD_PRELOAD=/tmp/evil.so，劫持 libc 函数

2、LD_LIBRARY_PATH
作用：动态链接库搜索路径，可被用于加载恶意库文件
攻击场景: 将恶意库路径添加到 LD_LIBRARY_PATH，程序会优先加载恶意库
"LD_LIBRARY_PATH",

3、PROMPT_COMMAND
作用：Bash 提示符命令，每次显示提示符时都会执行
攻击场景: 设置 PROMPT_COMMAND="malicious_command"，每次命令执行前都会运行恶意代码
"PROMPT_COMMAND",

4、PS1
作用：Bash 提示符变量，虽然主要用于显示，但可能被用于隐藏恶意命令
攻击场景: 在 PS1 中嵌入命令执行，每次显示提示符时执行
"PS1",

5、PATH
作用：可执行文件搜索路径，可被用于路径劫持攻击
攻击场景: 将 /tmp 添加到 PATH 前面，放置同名恶意程序，劫持系统命令
"PATH",

6、HISTFILE
作用：命令历史文件路径，可被用于隐藏命令执行痕迹
攻击场景: 设置 HISTFILE=/dev/null，禁用命令历史记录，隐藏攻击痕迹
"HISTFILE",

7、HISTCONTROL
作用：命令历史控制，可被用于隐藏敏感命令
攻击场景: 设置 HISTCONTROL=ignorespace，以空格开头的命令不会被记录
"HISTCONTROL",

8、HTTP_PROXY/HTTPS_PROXY
作用：代理服务器设置，可被用于中间人攻击或数据泄露
攻击场景: 设置恶意代理服务器，拦截和窃取网络流量
"HTTP_PROXY",
"HTTPS_PROXY",

9、NO_PROXY
作用：代理排除列表，可能被用于绕过安全检测
攻击场景: 将敏感域名添加到 NO_PROXY，绕过代理监控
"NO_PROXY",

10、TMPDIR/TMP/TEMP
作用：临时目录路径，可被用于放置恶意文件或隐藏攻击载荷
攻击场景: 设置 TMPDIR=/tmp/evil，将恶意文件放在非标准位置，逃避检测
"TMPDIR",
"TMP",
"TEMP",