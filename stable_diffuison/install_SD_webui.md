# Stable Diffusion WEB UI Installation Guide
## 前言
### 1、为什么要本地部署
因为没有生成数量的限制，不用花钱，不用被nsfw约束，生成时间快，不用排队，自由度高很多，可以调试和个性化的地方也更多。

### 2、本地化部署的要求
本地化部署运行虽然很好，但是也有一些基本要求
- 需要拥有NVIDIA显卡，GT1060起，显存4G以上。（已经不需要3080起，亲民不少）
- 操作系统需要win10或者win11的系统。
- 电脑内存16G或者以上。
- 最好会魔法上网，否则网络波动，有些网页打不开，有时下载很慢。

### 使用的项目Stable diffusion WebUI项目

Stable diffusion大家都知道了，是当前最多人使用且效果最好的开源AI绘图软件之一，属于当红炸子鸡了。

不过，stable diffusion项目本地化的部署，是纯代码界面，使用起来对于非程序员没那么友好。

而stable diffusion webui，是基于stable diffusion 项目的可视化操作项目。

通过可视化的网页操作，更方便调试prompt，及各种参数。

## 电脑环境配置
### 1.安装miniconda
这个是用来管理python版本的，他可以实现python的多版本切换。

下载地址：https://link.zhihu.com/?target=http%3A//docs.conda.io/en/latest/miniconda.html

安装时按默认的一路next就行。

### 2.打开miniconda，输入conda -V 弹出版本号即为正确安装

### 3.配置库包下载环境，加快网络速度（替换下载库包地址为国内的清华镜像站）

```shell
conda config --set show_channel_urls yes 
```

生成.condarc 文件

在我的电脑/此电脑-C盘-users-你的账号名下用记事本打开并修改.condarc文件。（如我的路径是C:\Users\Administrator。）

把下面的内容全部复制进去，全部覆盖原内容，ctrl+s保存，关闭文件。

```yaml

channels:
 - defaults
show_channel_urls: true
default_channels:
 - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/main
 - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/r
 - https://mirrors.tuna.tsinghua.edu.cn/anaconda/pkgs/msys2
custom_channels:
 conda-forge: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 msys2: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 bioconda: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 menpo: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 pytorch: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 pytorch-lts: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
 simpleitk: https://mirrors.tuna.tsinghua.edu.cn/anaconda/cloud
```

运行conda clean -i 清除索引缓存，以确保使用的是镜像站的地址。

### 4.切换成其他盘来创建python环境

如果继续操作，会把整个项目创建在c盘，而很多人c盘容量紧张，可以创建在其他盘，比如D盘。

输入D: 然后回车。

（后来才发现这一步并不能把项目装在d盘，他仍然是在c盘，不过没关系，他很小，不会占用太多空间，那咱继续往下操作）

### 5.创建python 3.10.6版本的环境

运行下面语句，创建环境。

```shell

conda create --name stable-diffusion-webui python=3.10.6
```

系统可能会提示y/n, 输入y，按回车即可。显示done，那就完成了。

### 6.激活环境

输入conda activate stable-diffusion-webui 回车。

### 7.升级pip，并设置pip的默认库包下载地址为清华镜像。

每一行输入后回车，等执行完再输入下一行，再回车

```shell

python -m pip install --upgrade pip
pip config set global.index-url https://pypi.tuna.tsinghua.edu.cn/simple
```

### 8.安装git，用来克隆下载github的项目，比如本作中的stable diffusion webui
前往git官网 https://link.zhihu.com/?target=http%3A//git-scm.com/download/win

下载好后，一路默认安装，next即可。

开始菜单-输入“git”，找到git cmd。

打开并输入下面指令。

## 安装cuda
cuda是NVIDIA显卡用来跑算法的依赖程序，所以我们需要它。

打开NVIDIA cuda官网，http://developer.nvidia.com/cuda-toolkit-archive

（这里有人可能会打不开网页，如果打不开，请用魔法上网。）

你会发现有很多版本，下载哪个版本呢？


回到一开始的miniconda的小窗，输入nvidia-smi，查看你的cuda版本


比如我的是11.7的版本，我就下载11.7.0的链接，


然后按照自己的系统，选择win10或者11，exe local，download

下载完后安装，这个软件2个G，可以安装在c盘以外的地方。比如D盘。

好了，完成这步，电脑的基础环境设置终于完事了。

下面开始正式折腾stable diffusion了。

## stable diffusion环境配置
1.下载stable diffusion源码

确认你的miniconda黑色小窗显示的是

（stable-diffusion-weibui）D:\>
如果不是，则输入D: 按回车。

（当然你也可以放在其他你想放的盘的根目录里面。

不建议放在c盘，因为这个项目里面有一些模型包，都是几个G几个G的，很容易你的C盘就满了，其他盘容量在10G以上的就都行。

放其他盘，则输入比如e: f: g: 等，然后回车即可。）

再来克隆stable diffusion webui项目（下面简称sd-webui）

接着执行

git clone https://github.com/AUTOMATIC1111/stable-diffusion-webui.git
直到显示done即可。

注意，现在克隆的本地地址，就是下面经常提到的“项目根目录”。比如，我的项目根目录是D:\stable-diffusion-webui

### 2.下载stable diffusion的训练模型
在http://huggingface.co/CompVis/stable-diffusion-v-1-4-original/tree/main

点击file and versions选项卡，下载sd-v1-4.ckpt训练模型。

（需要注册且同意协议，注册并同意协议之后即可下载）


注：这个模型是用于后续生成AI绘图的绘图元素基础模型库。

后面如果要用waifuai或者novelai，其实更换模型放进去sd-webui项目的模型文件夹即可。

我们现在先用stable diffusion 1.4的模型来继续往下走。



3.下载好之后，请把模型更名成model.ckpt,然后放置在sd-webui的models/stable-diffusion目录下。比如我的路径是D:\stable-diffusion-webui\models\Stable-diffusion


4. 安装GFPGAN

这是腾讯旗下的一个开源项目，可以用于修复和绘制人脸，减少stable diffusion人脸的绘制扭曲变形问题。


打开http://github.com/TencentARC/GFPGAN

把网页往下拉，拉到readme.md部分，找到V1.4 model，点击蓝色的1.4就可以下载。


下载好之后，放在sd-webui项目的根目录下面即可，比如我的根目录是D:\stable-diffusion-webui



4.在miniconda的黑色小窗，准备开启运行ai绘图程序sd-webui

输入

cd stable-diffusion-webui
进入项目的根目录。

如果你安装在其他地方，也是同理，按下面的方法去进入项目根目录。

输入盘符名称加上冒号（如c: d: e:）即可进入磁盘根目录。

输入cd..即可退出至上一级目录，

输入cd + abc即可进入abc文件夹。（如cd stable-diffusion-webui，前提是你有相应的文件夹，否则会报错）

总之，要进入sd-webui的项目根目录后，才能执行下面的指令，否则会报错。

（这个根目录是上面git clone 指令时候创建的stable-diffusion-webui根目录，不是在c盘miniconda里面的那个stable-diffusion-webui根目录。）

接着执行

webui-user.bat
然后回车，等待系统自动开始执行。

直到系统提示，running on local URL: http://127.0.0.1:7860

这就代表，你可以开始正式使用AI画画啦~


注意：

这一步可能经常各种报错，需要耐心和时间多次尝试。

不要关闭黑色小窗，哪怕它几分钟没有任何变化。

如果提示连接错误，可能需要开启或者关闭魔法上网，再重新执行webui-user.bat命令。

如果不小心退出了黑色窗口，则重新点击：开始菜单-程序-打开miniconda窗口，输入

conda activate stable-diffusion-webui
并进入sd-webui项目根目录再执行

webui-user.bat

## 一些问题的解决办法
### webui-user.bat 下载组件时ssl error
将launch.py中有关github的地址前加上加速地址如：
```python
gfpgan_package = os.environ.get('GFPGAN_PACKAGE', "git+https://ghproxy.com/https://github.com/TencentARC/GFPGAN.git@8d2447a2d918f8eba5a4a01463fd48e45126a379")
```

### webui-user.bat 启动时一直没动静
在launch.py 的138行 run_pip方法中
```python
 index_url_line = f' --index-url {index_url}' if index_url != '' else ''
```
后增加 
```python
print({python})
print({command})
print({index_url_line})
```
然后再执行webui-user.bat，这个时候会打印出接下来将要执行的pip命令，在cmd中手动执行该命令就可以看到详细的执行进度及错误信息。


## 模型资源网站
[civitai(需要魔法)](https://civitai.com/)
[cyberes](https://cyberes.github.io/stable-diffusion-models/)
[publicprompts](https://publicprompts.art/)
[huggingface](https://huggingface.co/)
[aimodel](https://aimodel.subrecovery.top/)
[123114514](https://www.123114514.xyz/)