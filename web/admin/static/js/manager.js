;!function (win) {
    "use strict";

    $.ajaxSetup({
        headers: {
            'JWT': localStorage.getItem('jwt'),
        },
        crossDomain: true,
        xhrFields: {withCredentials: true},
        complete: function (XMLHttpRequest, textStatus) {

            if (XMLHttpRequest.status === 500) {
                layer.msg('服务器内部错误');
            }

            if (XMLHttpRequest.responseJSON.code === 401) {
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

    win.Admin = function (api) {
        this.api = api;
        this.path = '';
        this.helper = new helper();

        let admin = this;

        this.setPath = function (path) {
            this.path = path;
        };

        this.send = function (method, data) {
            let url = admin.api + admin.path;
            if (method.toUpperCase() === 'GET') {
                url = url + admin.path;
            }

            let put = '';
            if (method.toUpperCase() === 'DELETE') {
                if (data.hasOwnProperty('id')) {
                    put = '?id=' + data['id'];
                } else {
                    put = '?_id=' + data['_id'];
                }
            }

            $.ajax({
                url: url + put,
                method: method,
                data: data,
                success: function (data) {
                    layer.msg('操作成功', {icon: 1, time: 1000}, function () {
                        layer.closeAll();
                        layui.table.reload('data-list', {
                            url: admin.api + admin.path + '/list',
                        });
                    });
                },
                error: function (data) {
                    console.log(data);
                }
            })
        };

        this.newTable = function (path, cols) {
            this.setPath(path);
            let table = new this.Table(cols);
            table.render();
            return table;
        };

        this.login = function () {
            let form = layui.form;
            form.on('submit(login)', function (data) {
                console.log(data.field["username"], data.field["password"]);
                localStorage.setItem("jwt", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJJZCI6IjVkNGJjNDFhODBhMWNkNDAwYWU3MTVmNyIsIkNoZWNrRGF0YSI6IntcIndlYlwiOjJ9IiwiZXhwIjoxNTY2ODA5NTA2fQ.L2rugS24FNXIJ73cHTc-VaSZsCx60r7FPbOSddFU0_o");
                location.href = "/main";
                return false;
            });
        };

        this.logout = function () {
            localStorage.removeItem("jwt");
            location.href = "/login";
        };

        this.Form = function () {
            let form = this;
            this.editors = {};
            this.isEdit = false;
            this.fields = [];

            this.formArrayFixed = function (data) {
                for (let i in data) {
                    if (data.hasOwnProperty(i)) {
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

                for (let i in form.editors) {
                    if (form.editors.hasOwnProperty(i)) {
                        data[i] = form.editors[i].txt.html()
                    }
                }

                return data;
            };

            this.dataFilter = function (formData, editData) {
                let result = {};

                let hasChange = function (formDataVal, editDataVal) {
                    if ((typeof editDataVal) === "object") {
                        if (Array.isArray(editDataVal)) {
                            return JSON.stringify(editDataVal.sort()) !== JSON.stringify(formDataVal.sort());
                        } else {
                            return !admin.helper.objectEquals(formDataVal, editDataVal);
                        }
                    } else {
                        return formDataVal != editDataVal
                    }
                };

                for (let i in formData) {
                    console.log(formData[i], editData[i], i, hasChange(formData[i], editData[i]));
                    if (hasChange(formData[i], editData[i])) {
                        result[i] = formData[i];
                    }
                }

                if (editData.hasOwnProperty('_id')) {
                    result['_id'] = editData['_id'];
                }

                if (editData.hasOwnProperty('id')) {
                    result['id'] = editData['id'];
                }

                return result;

            };

            this.open = function (data) {
                console.log(data);
                this.isEdit = data !== undefined;
                layer.open({
                    type: 1,
                    shade: false,
                    maxmin: true,
                    title: "编辑/添加",
                    content: document.getElementById('edit-form').innerHTML,
                    zIndex: layer.zIndex,
                    success: function (layerIndex) {
                        layer.setTop(layerIndex);
                    }
                });
                layui.use(['form', 'upload'], function () {
                    let upload = layui.upload;
                    layui.form.on('submit(form-submit)', function (formData) {
                        if (form.isEdit) {
                            admin.send('put', form.dataFilter(form.formArrayFixed(formData.field), data));
                        } else {
                            admin.send('post', form.formArrayFixed(formData.field));
                        }
                        return false;
                    });
                    form.checkbox();
                    form.radio();
                    form.richText();
                    form.upload(upload);
                    layui.form.val('edit-form', data);
                    console.log('complete');
                    layui.form.render();
                    form.fill(data);
                    layui.form.render();
                });
            };

            this.fill = function (data) {
                if (data === undefined)
                    return;
                for (let i = 0; i < form.fields.length; i++) {
                    switch (form.fields[i].type) {
                        case 'checkbox':
                        case 'radio':
                            let vals = data[form.fields[i].name];
                            if (Array.isArray(vals)) {
                                for (let index = 0; index < vals.length; index++) {
                                    $('input[value="' + vals[index] + '"]').attr("checked", true);
                                }
                            } else {
                                $('input[value="' + vals + '"]').attr("checked", true);
                            }
                            break;
                        case 'upload':
                            let val = data[form.fields[i].name];
                            if ( val === null || val === "" )
                                break;
                            if (Array.isArray(val)) {
                                for (let index = 0; index < val.length; index++) {
                                    $('#upload-file-' + form.fields[i].name).append("<div style='float: left' onclick='admin.helper.deleteSelf(this)'>" +
                                        "<input type='hidden' value='" + val[index] + "' name='" + form.fields[i].name + "[]'>" +
                                        "<img style='margin-right:10px; width: 50px; height: 50px' alt src='" + val[index] + "'>" +
                                        "</div>");
                                }
                            } else {
                                $('#upload-file-' + form.fields[i].name).append("<div style='float: left' onclick='admin.helper.deleteSelf(this)'>" +
                                    "<input type='hidden' value='" + val + "' name='" + form.fields[i].name + "'>" +
                                    "<img style='margin-right:10px; width: 50px; height: 50px' alt src='" + val + "'>" +
                                    "</div>");
                            }

                            break;
                    }
                }

                console.log(this.editors);
                for (let i in this.editors) {
                    if (this.editors.hasOwnProperty(i)) {
                        this.editors[i].txt.html(data[i]);
                    }
                }
            };

            /**
             * 复选框数据生成
             */
            this.checkbox = function () {
                let checkboxes = document.getElementsByClassName('checkbox-list');
                for (let index = 0; index < checkboxes.length; index++) {
                    let element = checkboxes[index];
                    let data = element.getAttribute('data-list'), name = element.getAttribute('data-name'),
                        key = element.getAttribute('data-key'), val = element.getAttribute('data-val'),
                        input = document.getElementById('form-' + name);

                    try {
                        // 将字符串作为
                        let json = eval(data);
                        for (let i = 0; i < json.length; i++) {
                            input.innerHTML += '<input value="' + json[i][key] + '" type="checkbox" name="' + name + '[]" title="' + json[i][val] + '">'
                            this.fields.push({name: '' + name + '', val: json[i][key], type: 'checkbox'});
                        }
                    } catch (e) {
                        console.log('是数据地址, 将从api服务器获取数据');
                        layui.$.ajax({
                            url: admin.api + data,
                            async: false,
                            success: function (data) {
                                for (let i = 0; i < data.data.length; i++) {
                                    input.innerHTML += '<input value="' + data.data[i][key] + '" type="checkbox" name="' + name + '[]" title="' + data.data[i][val] + '">'
                                    form.fields.push({
                                        name: '' + name + '',
                                        val: data.data[i][key],
                                        type: 'checkbox'
                                    });
                                }
                            }
                        });
                    }
                }
            };

            /**
             * 富文本输入框生成
             */
            this.richText = function () {
                let textareas = $('.textarea'), editor = wangEditor;

                for (let i = 0; i < textareas.length; i++) {
                    form.editors[$(textareas[i])[0].getAttribute('name')] = new editor($(textareas[i])[0]);
                    form.editors[$(textareas[i])[0].getAttribute('name')].create();
                }
            };

            /**
             * 文件上传组件
             * @param upload
             */
            this.upload = function (upload) {
                let uploads = $('.form-upload');
                for (let i = 0; i < uploads.length; i++) {
                    let multiple = $(uploads[i]).attr("data-multiple");
                    form.fields.push({
                        name: $(uploads[i]).attr("data-key"),
                        type: 'upload'
                    });
                    upload.render({
                        elem: $(uploads[i])[0],
                        url: admin.api + '/upload',
                        headers: {JWT: localStorage.getItem('jwt')},
                        multiple: multiple === "true",
                        data: {key: $(uploads[i]).attr("data-key")},
                        done: function (data) {
                            for (let key in data.data) {
                                if (data.data.hasOwnProperty(key)) {
                                    let name = key;
                                    if (multiple === "true") {
                                        name = key + "[]";
                                    } else {
                                        $('#upload-file-' + key).html("");
                                    }
                                    for (let index = 0; index < data.data[key].length; index++) {

                                        $('#upload-file-' + key).append("<div style='float: left' onclick='admin.helper.deleteSelf(this)'>" +
                                            "<input type='hidden' value='" + data.data[key][index] + "' name='" + name + "'>" +
                                            "<img style='margin-right:10px; width: 50px; height: 50px' alt src='" + data.data[key][index] + "'>" +
                                            "</div>");
                                    }
                                }
                            }
                        }
                    });
                }
            };

            /**
             * 单选框数据生成
             */
            this.radio = function () {
                let radioList = document.getElementsByClassName('radio-list');
                for (let index = 0; index < radioList.length; index++) {
                    let element = radioList[index];
                    let data = element.getAttribute('data-list'), name = element.getAttribute('data-name'),
                        key = element.getAttribute('data-key'), val = element.getAttribute('data-val'),
                        input = document.getElementById('form-radio-' + name);
                    try {
                        // 将字符串作为
                        let json = eval(data);
                        for (let i = 0; i < json.length; i++) {
                            input.innerHTML += '<input value="' + json[i][key] + '" type="radio" name="' + name + '" title="' + json[i][val] + '">'
                            form.fields.push({
                                name: name,
                                val: json[i][key],
                                type: 'radio'
                            });
                        }
                    } catch (e) {
                        console.log('是数据地址, 将从api服务器获取数据');
                        // console.log(e);
                        layui.$.ajax({
                            url: admin.api + data,
                            async: false,
                            success: function (data) {
                                for (let i = 0; i < data.data.length; i++) {
                                    input.innerHTML += '<input value="' + data.data[i][key] + '" type="radio" name="' + name + '" title="' + data.data[i][val] + '">'
                                    form.fields.push({
                                        name: name,
                                        val: data.data[i][key],
                                        type: 'radio'
                                    });
                                }
                            }
                        });
                    }
                }
            }
        };

        /**
         * 数据操作
         * @param cols
         * @constructor
         */
        this.Table = function (cols) {
            let table = this;
            this.events = {};

            this.attribute = {
                elem: '#data-list',
                url: admin.api + admin.path + '/list',
                page: true,
                toolbar: true,
                autoSort: false,
                cols: cols,
                limitName: 'rows',
                response: {
                    msgName: 'message',
                },
                parseData: function (res) {
                    return res
                }
            };

            /**
             * 表单事件
             */
            this.event = function () {
                this.delete = function (data) {
                    layer.confirm('真的要删除吗?', function () {
                        console.log('准备删除');
                        console.log(data);
                        admin.send('delete', data);
                        layer.close(layer.index);
                    });
                };

                this.edit = function (data) {
                    let form = new admin.Form();
                    form.open(data);
                }
            };

            this.sort = function (obj) {
                if (obj.type === 'desc') {
                    return "-" + obj.field;
                } else {
                    return obj.field;
                }
            };

            this.date2unix = function (date) {
                date = date.trim();
                date = date.substring(0, 19);
                date = date.replace(/-/g, '/');
                return new Date(date).getTime() / 1000;
            };

            /**
             * 搜索操作
             */
            this.search = function () {
                layui.use('form', function () {
                    let form = layui.form;
                    form.on('submit(search)', function (data) {
                        let section = 'section=';
                        for (let key in data.field) {
                            if (data.field.hasOwnProperty(key)) {
                                if (key.startsWith("section:")) {
                                    let realKey = key.replace("section:", "");
                                    let date = data.field[key].split('~');
                                    if (date.length === 2) {
                                        // 如果是一个时间区间, 生成 -key:date,date
                                        section += realKey + ':' + table.date2unix(date[0]) + ',' + table.date2unix(date[1]) + '&section=';
                                    } else {
                                        if (date[0] !== '' && date[0].length > 0) {
                                            // 如果是一个时间, 那么需要在输入框name设置为 section:-key:date
                                            section += realKey + ':' + table.date2unix(date[0]) + '&section=';
                                        }
                                    }
                                    delete data.field[key];
                                }
                            }
                        }

                        section = section.substring(0, section.length - '&section='.length);
                        layui.table.reload('data-list', {
                            url: admin.api + admin.path + '/list?' + section,
                            where: data.field
                        });
                        return false;
                    });
                });
            };

            this.render = function () {
                let inputs = document.getElementsByClassName('search-date-pick');
                for (let i = 0; i < inputs.length; i++) {
                    admin.helper.datepick('#' + inputs[i].id, inputs[i].getAttribute("data-range"), inputs[i].getAttribute("data-type"));
                }

                layui.use('table', function () {
                    layui.table.render(table.attribute);
                    // 自定义排序以符合后端接口规则
                    layui.table.on('sort(data-list)', function (obj) {
                        layui.table.reload('data-list', {initSort: obj, where: {sorts: table.sort(obj)}});
                    });
                });

                this.curd();
                this.search();
            };

            this.curd = function () {
                layui.use('table', function () {
                    layui.table.on('tool(data-list)', function (obj) {
                        let data = obj.data;
                        let event = obj.event;
                        let tr = obj.tr;

                        if (table.events.hasOwnProperty(event)) {
                            table.events[event](tr, data, event);
                        } else {
                            let fn = new table.event();
                            if (fn.hasOwnProperty(event)) {
                                fn[event](data);
                            }
                        }
                    });
                });
            }

        };
    };

    /**
     *
     */
    win.helper = function () {
        /**
         * 金钱 分 转为 元
         * @param data
         * @returns {number}
         */
        this.fen2yuan = function (data) {
            return (data[this.field] / 100)
        };

        /**
         * 时间戳转时间
         * @param data
         * @returns {string}
         */
        this.time2str = function (data) {
            return new Date(data[this.field] * 1000).toLocaleDateString() + ' ' + new Date(data[this.field] * 1000).toLocaleTimeString()
        };

        /**
         * 生成一个时间选择器
         * @param id
         * @param range 是否可以区间选择,布尔值
         * @param type 可选值可参考 layui 时间选择器文档,默认 datetime
         */
        this.datepick = function (id, range, type) {
            range = range !== false && range !== '' ? '~' : '';
            type = type === undefined || type === '' ? 'datetime' : type;

            layui.use('laydate', function () {
                layui.laydate.render({elem: id, range: range, type: type});
            });
        };

        this.deleteSelf = function (self) {
            $(self).remove();
        };

        this.objectEquals = function (x, y) {
            console.log(11);
            if (x === y) {
                return true;
            }

            if (!(x instanceof Object) || !(y instanceof Object)) {
                return false;
            }

            if (x.constructor !== y.constructor) {
                return false;
            }
            for (let p in x) {
                if (x.hasOwnProperty(p)) {
                    if (!y.hasOwnProperty(p)) {
                        return false;
                    }
                    if (x[p] === y[p]) {
                        continue;
                    }
                    if (typeof (x[p]) !== "object") {
                        return false;
                    }
                    if (!Object.equals(x[p], y[p])) {
                        return false;
                    }
                }
            }
            return true;
        };

    };
}(window);