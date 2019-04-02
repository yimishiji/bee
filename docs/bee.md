

相对于原生bee, curd
===

## 配置 bee 脚本配置
src/appname/bee.json
```$xslt
    {
        "version": 0,
        "database": {
            "driver": "mysql",
            "conn": "username:password(localhost:3306)/dbname",
        },
        "cmd_args": [],
        "enable_reload": true
    }
```


##  生成app代码
项目目录下执行以下代码将生成api-vue代码
```$xslt
 bee generate appcode -tables="tablename" -level=4
```

