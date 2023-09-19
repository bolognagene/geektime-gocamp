import React from 'react';
import { Button, Form, Input } from 'antd';
import axios from "@/axios/axios";
import Link from "next/link";
import router from "next/router";

const onFinish = (values: any) => {
    axios.post("/users/login", values)
        .then((res) => {
            if(res.status != 200) {
                alert(res.statusText);
                return
            }
            alert(res.data)
            router.push('/users/profile')
        }).catch((err) => {
            alert(err);
    })
};

const onFinishFailed = (errorInfo: any) => {
    alert("输入有误")
};

const LoginForm: React.FC = () => {
    return (<Form
        name="basic"
        labelCol={{ span: 8 }}
        wrapperCol={{ span: 16 }}
        style={{ maxWidth: 600 }}
        initialValues={{ remember: true }}
        onFinish={onFinish}
        onFinishFailed={onFinishFailed}
        autoComplete="off"
    >
        <Form.Item
            label="邮箱"
            name="email"
            rules={[{ required: true, message: '请输入邮箱' }]}
        >
            <Input />
        </Form.Item>

        <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
        >
            <Input.Password />
        </Form.Item>

        <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
            <Button type="primary" htmlType="submit">
                登录
            </Button>
            <Link href={"/users/login_sms"} >
                &nbsp;&nbsp;手机号登录
            </Link>
            <Link href={"/users/login_wechat"} >
                &nbsp;&nbsp;微信扫码登录
            </Link>
            <Link href={"https://open.weixin.qq.com/connect/qrconnect?appid=wx1b4c7610fc671845&redirect_uri=https%3A%2F%2Faccount.geekbang.org%2Faccount%2Foauth%2Fcallback%3Ftype%3Dwechat%26ident%3D124963%26login%3D0%26cip%3D0%26redirect%3Dhttps%253A%252F%252Faccount.geekbang.org%252Fthirdlogin%253Fremember%253D1%2526type%253Dwechat%2526is_bind%253D0%2526platform%253Dtime%2526redirect%253Dhttps%25253A%25252F%25252Ftime.geekbang.org%25252F%2526failedurl%253Dhttps%253A%252F%252Faccount.geekbang.org%252Fsignin%253Fredirect%253Dhttps%25253A%25252F%25252Ftime.geekbang.org%25252F&response_type=code&scope=snsapi_login&state=0674438c3faded355c5e8eac545681c1#wechat_redirect"} >
                &nbsp;&nbsp;极客时间微信扫码登录V1
            </Link>
            <Link href={"/users/signup"} >
                &nbsp;&nbsp;注册
            </Link>
        </Form.Item>
    </Form>
)};

export default LoginForm;