;!function (win) {
    "use strict";
    /**
     * 后台管理
     * @param api
     * @constructor
     */
    win.Admin = function (api) {
        this.api = api;
        this.jwt = localStorage.getItem('jwt');
        this.editors = {};
    };

    Admin.prototype.jquerySetup = function () {
        $.ajaxSetup({
            complete: function (XMLHttpRequest, textStatus) {

                switch (XMLHttpRequest.status) {
                    case 500:
                        layer.msg('服务器内部错误');
                }

                switch (XMLHttpRequest.responseJSON.code) {
                    case 401:
                        layer.msg(XMLHttpRequest.responseJSON.message, function () {
                            if (window !== top) {
                                top.location.href = '/login';
                            } else {
                                location.href = '/login';
                            }
                        })
                }
            }
        });
    };

    /**
     * 登录
     * @param {string} username
     * @param {string} password
     */
    Admin.prototype.login = function (username, password) {
        // 登录逻辑
        console.log(this.api);
        localStorage.setItem("jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJZCI6IjVkNGJjNDFhODBhMWNkNDAwYWU3MTVmNyIsIkxvZ2luQXQiOjEsImV4cCI6MTU2NjU2MjkzNn0.qJEqqh0XPNStai2jPdVO2wAD2Fpir5hslgX0WblyQYU");
        location.href = "/main";
    };

    Admin.prototype.deleteData = function (data) {
        let _this = this;
        let putData = '';
        if (data.hasOwnProperty('id')) {
            putData = 'id=' + data['id'];
        } else {
            putData = '_id=' + data['_id'];
        }
        layui.$.ajax({
            url: _this.api + _this.path + '?' + putData,
            method: 'DELETE',
            headers: {
                'JWT': _this.jwt,
            },
            crossDomain: true,                 //加这二行支持ajax跨域，允许跨域
            xhrFields: {withCredentials: true},//加这二行支持ajax跨域，携带凭证
            success: function (data) {
                layui.table.reload('data-list', {
                    url: admin.api + _this.path + '/list',
                });
            },
            error: function (data) {
                console.log(data);
            }
        });
        layer.close(layer.index);
    };

    Admin.prototype.addData = function (data, method) {
        let _this = this;
        layui.$.ajax({
            url: _this.api + _this.path,
            method: method,
            data: _this.formArrayFixed(data),
            headers: {
                'JWT': _this.jwt,
            },
            crossDomain: true,                 //加这二行支持ajax跨域，允许跨域
            xhrFields: {withCredentials: true},//加这二行支持ajax跨域，携带凭证
            success: function (data) {
                layui.table.reload('data-list', {
                    url: admin.api + _this.path + '/list',
                });
            },
            error: function (data) {
                console.log(data);
            }
        });
        layer.close(layer.index);
    };

    Admin.prototype.formArrayFixed = function (data) {
        for (let i in data) {
            if ( data.hasOwnProperty(i) ) {
                if (/\w+\[\d+]/.test(i)) {
                    let key = i.match(/(\w+)\[\d+]/)[1];
                    if (data.hasOwnProperty(key)) {
                        data[key].push(data[i])
                    } else {
                        data[key] = [];
                        data[key].push(data[i]);
                    }
                    delete data[i];
                }
            }
        }

        for ( let i in this.editors ) {
            if (this.editors.hasOwnProperty(i)) {
                data[i] = this.editors[i].txt.html()
            }
        }

        return data;
    };

    Admin.prototype.openForm = function (editData) {
        let _this = this;
        layer.open({
            type: 1,
            shade: false,
            maxmin: true,
            title: "编辑/添加",
            content: document.getElementById('edit-form').innerHTML,
            zIndex: layer.zIndex, //重点1
            success: function (layero) {
                layer.setTop(layero); //重点2
            }
        });
        // layer.open({
        //     type: 1,
        //     title: "编辑/添加",
        //     content: document.getElementById('edit-form').innerHTML,
        //     zIndex: layer.zIndex,
        //     success: function (layero) {
        //         layer.full();
        //         layer.min();
        //         layer.restore();
        //         layer.setTop(layero);
        //     }
        // });

        let dataValid = function (formData, data) {
            let result = {};
            for (let i in formData) {
                if (data[i] != formData[i]) {
                    result[i] = formData[i];
                }
            }

            if (data.hasOwnProperty('_id')) {
                result['_id'] = data['_id'];
            }

            if (data.hasOwnProperty('id')) {
                result['id'] = data['id'];
            }

            return result;
        };

        layui.use('form', function () {
            let form = layui.form;

            let checkboxList = document.getElementsByClassName('checkbox-list');
            for (let index = 0; index < checkboxList.length; index++) {
                let element = checkboxList[index];
                let data = element.getAttribute('data-list'),
                    name = element.getAttribute('data-name'),
                    key = element.getAttribute('data-key'),
                    val = element.getAttribute('data-val'),
                    input = document.getElementById('form-' + name);
                try {
                    // 将字符串作为
                    let json = eval(data);
                    for (let i = 0; i < json.length; i++) {
                        input.innerHTML += '<input value="' + json[i][key] + '" type="checkbox" name="' + name + '[]" title="' + json[i][val] + '">'
                    }
                    // form.on('submit(form-submit)', function (formData) {
                    //     if (editData !== null && editData !== undefined) {
                    //         _this.addData(dataValid(formData.field, editData), 'PUT');
                    //         return false;
                    //     }
                    //     _this.addData(formData.field, 'POST');
                    //     return false;
                    // });
                    // if (editData !== null && editData !== undefined) {
                    //     form.val('edit-form', editData);
                    // }
                    // layui.use('layedit', function () {
                    //     let layedit = layui.layedit;
                    //     let index = layedit.build('textarea'); //建立编辑器
                    //     layedit.sync(index)
                    // });
                    // form.render();
                } catch (e) {
                    console.log('是数据地址, 将从api服务器获取数据');
                    // console.log(e);
                    layui.$.ajax({
                        url: _this.api + data,
                        headers: {
                            'JWT': _this.jwt,
                        },
                        async: false,
                        crossDomain: true,                 //加这二行支持ajax跨域，允许跨域
                        xhrFields: {withCredentials: true},//加这二行支持ajax跨域，携带凭证
                        success: function (data) {
                            for (let i = 0; i < data.data.length; i++) {
                                console.log(data.data[i]);
                                input.innerHTML += '<input value="' + data.data[i][key] + '" type="checkbox" name="' + name + '[]" title="' + data.data[i][val] + '">'
                            }
                        }
                    });
                }
            }

            form.on('submit(form-submit)', function (formData) {
                if (editData !== null && editData !== undefined) {
                    _this.addData(dataValid(formData.field, editData), 'PUT');
                    return false;
                }
                _this.addData(formData.field, 'POST');
                return false;
            });
            if (editData !== null && editData !== undefined) {
                form.val('edit-form', editData);
                for (let i in editData) {
                    if (editData.hasOwnProperty(i)) {
                        if ((typeof editData[i]) !== "string") {
                            for (let key in editData[i]) {
                                if (editData[i].hasOwnProperty(key))
                                    $('input[value="' + editData[i][key] + '"]').next().click();
                            }
                        }
                    }
                }
            }

            let textareas = $('.textarea');
            for (let i = 0; i < textareas.length; i++) {
                _this.editors[$(textareas[i])[0].getAttribute('name')] = new wangEditor($(textareas[i])[0]);
                _this.editors[$(textareas[i])[0].getAttribute('name')].create();
                if ( ((editData !== null && editData !== undefined) ) && editData.hasOwnProperty($(textareas[i])[0].getAttribute('name')) ) {
                    _this.editors[$(textareas[i])[0].getAttribute('name')].txt.html(editData[$(textareas[i])[0].getAttribute('name')]);
                }
            }

            form.render();
        })
    };

    /**
     * 退出
     */
    Admin.prototype.logout = function () {
        localStorage.removeItem("jwt");
        location.href = "/login";
    };

    /**
     * 单条数据的修改删除操作
     */
    Admin.prototype.curd = function () {
        let _this = this;
        layui.use('table', function () {
            let table = layui.table;
            table.on('tool(data-list)', function (obj) {
                let data = obj.data;
                let event = obj.event;
                let tr = obj.tr;

                let edit = function (data) {
                    console.log('准备编辑');
                    console.log(data);
                    _this.openForm(data);
                };

                let del = function (data) {
                    layer.confirm('真的要删除吗?', function () {
                        console.log('准备删除');
                        console.log(data);
                        _this.deleteData(data);
                        layer.close(layer.index);
                    });
                };

                switch (event) {
                    case 'edit':
                        edit(data);
                        break;
                    case 'delete':
                        del(data);
                        break;
                    default:
                        layui.msg('未找到对应操作事件');
                }
            });
        });
    };

    /**
     * 搜索操作
     */
    Admin.prototype.search = function () {
        let _this = this;
        layui.use('form', function () {
            let form = layui.form;
            form.on('submit(search)', function (data) {

                let dateToDate = function (date) {
                    date = date.trim();
                    date = date.substring(0, 19);
                    date = date.replace(/-/g, '/');
                    return new Date(date).getTime() / 1000;
                };

                let section = 'section=';
                for (let key in data.field) {
                    if (data.field.hasOwnProperty(key)) {
                        if (key.startsWith("section:")) {
                            let realKey = key.replace("section:", "");
                            let date = data.field[key].split('~');
                            if (date.length === 2) {
                                // 如果是一个时间区间, 生成 -key:date,date
                                section += realKey + ':' + dateToDate(date[0]) + ',' + dateToDate(date[1]) + '&section=';
                            } else {
                                if (date[0] !== '' && date[0].length > 0) {
                                    // 如果是一个时间, 那么需要在输入框name设置为 section:-key:date
                                    section += realKey + ':' + dateToDate(date[0]) + '&section=';
                                }
                            }
                            delete data.field[key];
                        }
                    }
                }

                section = section.substring(0, section.length - '&section='.length);
                layui.table.reload('data-list', {
                    url: admin.api + _this.path + '/list?' + section,
                    where: data.field
                });
                return false;
            });
        });
    };

    /**
     * 生成一个时间选择器
     * @param id
     * @param range 是否可以区间选择,布尔值
     * @param type 可选值可参考 layui 时间选择器文档,默认 datetime
     */
    Admin.prototype.datePick = function (id, range, type) {
        if (range !== false && range !== "") {
            range = '~';
        }

        if (type === undefined || type === "") {
            type = 'datetime';
        }

        layui.use('laydate', function () {
            let laydate = layui.laydate;
            laydate.render({
                elem: id,
                range: range,
                type: type
            });
        });

        return this;
    };

    /**
     * 绘制数据列表
     * @param path
     * @param cols
     * @param toolbar
     */
    Admin.prototype.renderList = function (path, cols, toolbar) {
        let inputs = document.getElementsByClassName('search-date-pick');
        for (let i = 0; i < inputs.length; i++) {
            this.datePick('#' + inputs[i].id, inputs[i].getAttribute("data-range"), inputs[i].getAttribute("data-type"));
        }
        let _this = this;
        this.path = path;
        this.search();
        this.curd();
        layui.use('table', function () {
            _this.jquerySetup();
            let table = layui.table;
            table.render({
                elem: '#data-list',
                url: admin.api + path + '/list',
                page: true,
                toolbar: true,
                autoSort: false,
                headers: {JWT: admin.jwt},
                cols: cols,
                limitName: 'rows',
                xhrFields: {withCredentials: true},
                response: {
                    msgName: 'message',
                },
                parseData: function (res) {
                    console.log(res);
                    return res
                }
            });
            // 自定义排序以符合后端接口规则
            table.on('sort(data-list)', function (obj) {
                table.reload('data-list', {
                    initSort: obj
                    , where: {
                        sorts: function () {
                            if (obj.type === 'desc') {
                                return "-" + obj.field;
                            } else {
                                return obj.field;
                            }
                        }()
                    }
                });
            });
        });
    };

    Admin.prototype.time2str = function (data) {
        return new Date(data[this.field] * 1000).toLocaleDateString() + ' ' + new Date(data[this.field] * 1000).toLocaleTimeString()
    };

    Admin.prototype.fen2yuan = function (data) {
        return (data[this.field] / 100)
    }
}(window);