# AST 是什么？
抽象语法树(Abstract Syntax Tree)简称AST.
它是源代码语法结构的一种抽象表示。它以树状的形式表现编程语言的语法结构，树上的每个节点都表示源代码中的一种结构。

# AST 的作用
- 编辑器的错误提示、代码格式化、代码高亮、代码自动补全
- elint、pretiier对代码错误或风格的检查
- webpack通过babel转译javascript语法

# AST如何生成
js执行的第一步是读取js文件中的字符流，然后通过词法分析生成token，之后再通过语法分析生成AST，最后生成机器码执行。

整个解析过程主要分为以下两个步骤：
- 分词：将整个代码字符串分割成最小语法单元数组
- 语法分析：在分词基础上建立分析语法单元之间的关系

JS Parser 是 js 语法解析器，它可以将 js 源码转成 AST，常见的 Parser 有 esprima、traceur、acorn、shift 等。


