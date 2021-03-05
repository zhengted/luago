# 

#### 介绍

- 本项目使用Go语言实现了Lua虚拟机和编译器（Lua5.3）
- 其中编译器部分分为了词法分析器（Lexer，也叫扫描器）、语法分析器（Parser）以及编译器（Compiler）

#### 软件架构

##### LuaVM层

- VM层的主要作用是根据lua的prototype（原型）中的指令（参考parser中的insts）执行。像极了计算机中CPU执行指令，虚拟机（VM：virtual machine）因此得名（计算机原理）。

- **如何查看一个lua文件的prototype内容**，很遗憾本项目暂不支持此功能，可以通过Lua官网下载Lua 5.3，使用命令```luac -l -l test.lua``` 查看

- 以以下lua代码为例

  ```lua
  -- test.lua
  local a = 20
  local b = 45
  print("Hello,World!")
  print(a)
  ```

- 执行完```luac -l -l test.lua```后如下

  ![image-20210305160636036](https://i.loli.net/2021/03/05/BAia2fCuGdQHKWb.png)

##### 编译器

- 编译过程

  ![image-20210305160913292](https://i.loli.net/2021/03/05/ywabmPx8iQY4A53.png)

- VM处理的内容都是针对Codegen生成的ByteCodes
- 而编译器则是从Source开始处理

###### 词法分析器（扫描器 Lexer）

- 词法分析主要使用有限状态机（Finite-state Machine,FSM），以注释为例，对应的FSM如下
- ![image-20210305162640328](https://i.loli.net/2021/03/05/drLUQe1S9WpAZOs.png)

- 词法分析的结果是一系列的token，以上面的lua代码为例，生成的词法内容如下
- ![image-20210305162941985](https://i.loli.net/2021/03/05/ba6V23u4DJOPwo1.png)

###### 语法分析器（Parser）

- 词法分析的任务是将token序列解析成抽象语法树（AST,Abstract Syntax Tree）
- 需要解析的内容：块Block、语句Statement、表达式Expression
- 由于不管在Statement还是在Expression中都会有解析Block的操作，存在递归操作。**递归下降分析法**

###### 代码生成（code generator）

- 根据不同的表达式或者语句发送指定的命令到block中
- TODO

#### 安装教程

1. 新建一个lua文件（以test.lua为例）

2. ```shell
   go install luago
   ```

3. ```shell
   luago test.lua	
   ```

#### 使用说明

- 具体lua文件路径根据实际情况自行做出调整