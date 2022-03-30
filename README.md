# piondemo
a demo from webrtc by pion
本项目用于学习[pion](https://github.com/pion)项目的使用
signalserver 一个简陋的信令服务器，负责交换sdp/candidates消息
p2pcelient   客户端，通过和signalserver交流，获取对端sdp/candidates，后续通过ICE完成udp数据链路的点对点打通
