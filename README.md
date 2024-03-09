# AList-ollama
一个云盘管理及图像智能搜索综合系统

## 介绍
Alist 作为一个云盘整合工具，为用户提供了便捷的云端文件管理体验。为了进一步提升用户使用体验，本项目通过集成 GPT 技术，实现了对云盘中图片的智能搜索和描述生成。用户可以通过图像关键词搜索以及由 GPT 生成的图像描述，更高效地定位和管理云盘中的图片资源。通过ollama可以更加方便的选择自定义模型，或是使用自己训练的模型，同时通过api更好的同Alist相结合。

功能：
实现基于关键词的图像搜索功能。
利用 GPT 技术生成准确、富有描述性的图片描述。
通过图像描述，为用户提供更直观、便捷的文件搜索和管理方式。

特点：
Alist 云盘整合： 利用 Alist 提供的云盘整合功能，将用户的云端图片资源汇聚在一个统一的工具中，方便集中管理。
GPT 图像描述技术： 结合 GPT 技术，对云盘中的图片进行智能描述生成，为每张图片生成富有描述性的文本。
关键词搜索： 提供基于关键词的图像搜索功能，使用户能够通过关键词快速定位所需图片。

## 下载及配置
```bash
git clone https://github.com/slowmagic10/alist-ollama.git
```
### Alist的下载及配置
#### Alist源码及文档
<https://github.com/alist-org/alist/tree/main><br>
<https://alist.nn.ci/zh/><br>

**1.环境准备**<br>
首先，你需要一个有```git```，```nodejs```，```pnpm```，```golang>=1.20```，```gcc```的环境<br>

**2.构建前端**<br>
```bash 
cd alist-ollama/alist-web
```
执行 ```pnpm install``` && ```pnpm build``` 得到 dist 目录下的目标文件

**3.构建后端**<br>
```bash
cd ..
cd alist
```
将上一步的 ```dist``` 目录复制到项目下的 ```public``` 目录下，然后执行
```bash
bash run.sh
```

**4.运行服务**<br>
* 在Linux或Mac上运行
```bash
./alist server
```

* 在Windows上运行
```bash
./alist.exe server
```

**5.构建索引**<br>
登录后安装官方教程：```https://alist.nn.ci/zh/guide/drivers/common.html```挂载网盘<br>
进入管理界面，选择索引为云盘文件添加索引