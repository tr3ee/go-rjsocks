RJSocks 开源、自由、不受限制的Windows校园网认证客户端
==================================================

IEEE 802.1X EAPoL 校园网 (FZU) 认证客户端 for windows

#### 功能

1. 网络启用控制
2. 校园网认证
3. 支持客户端多点登录
4. 支持设备多网卡使用

#### 下载

下载 [最新版](https://github.com/tr3ee/go-rjsocks/releases)

#### 环境要求

需要安装 [WinPcap 4.1.0以上版本](https://www.winpcap.org/install/default.htm)

[Releases](https://github.com/tr3ee/go-rjsocks/releases) 页面中的RJSocks-Installer.exe自带WinPcap安装包，可以直接进行安装而跳过上述步骤

#### 简单使用

1. 填写用户名、密码并选择正确的网卡设备与网络适配器后，点击确定登录
2. 选择`记住密码`会把密码**明文**存放在 RJSocks.exe 目录下的 config.ini 文件中
3. 选择`自动登录`会再下一次打开时，跳过登录页直接登录
4. 在任务栏中可以找到 RJSocket 图标，右键弹出菜单
5. 图标说明：
    - ![Stop Icon](https://raw.githubusercontent.com/tr3ee/go-rjsocks/master/gui/resources/stop.ico) 表示当前认证状态是失败的，常见的场景有：错误的用户名密码，未勾选启用网络连接等
    - 白色火箭标识的RJSockets图标，则表示网络连接正常

#### 动态IP获取 (DHCP)

在一些特殊的场景中，RJSocks无法成功获取IP地址，可以通过图标右键菜单中的**刷新IP地址**手动刷新

#### 问题与反馈

任何意见、建议以及使用过程中的出现的问题，欢迎在 [Issues](https://github.com/tr3ee/go-rjsocks/issues) 提出

常见问题：

- 计算机待机或睡眠后无法联网，且图标为白色火箭标识的RJSockets图标
> 解决方案：
> 这是由于在待机或睡眠时网络设备会被关闭以节省电源，点击右键菜单中的**断开连接&重新认证**，重新认证即可

- 无法正常运行RJSocks，提示"无法打开配置文件config.ini"
> 解决方案：
> 这是由于RJSocks对当前目录没有读写权限导致的，可以通过修改文件夹`属性`中的`安全`，赋予当前用户读写权限

- 无法正常运行RJSocks，提示"无法打开日志文件log.txt"
> 解决方案：
> 这是由于RJSocks对当前目录没有读写权限导致的，可以通过修改文件夹`属性`中的`安全`，赋予当前用户读写权限

#### 开发与贡献

欢迎任何形式的参与和贡献，开发环境要求[Golang 1.9以上版本](https://golang.org/project/)并安装[GoPacket](https://github.com/google/gopacket)

贡献列表

| Name | GITHUB | Email |
|-| ------ | ---- |
| The Whisper | - | - |
| tr3e | https://github.com/tr3ee | tr3e.wang \[AT\] gmail.com |

联系作者Email：tr3e.wang \[AT\] gmail.com

# 许可证

本程序所有涉及的锐捷认证功能均为网上开源代码分析而得。
所有代码仅供学习交流使用，不得有意妨害任何一方权益。
一切使用后果由用户自己承担，以任何方式使用本程序即表示同意该声明。
除非另据说明，所有代码根据 [MIT 许可](https://github.com/tr3ee/go-rjsocks/edit/master/LICENSE) 发布。
