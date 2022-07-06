# go-chat 

### Dependencies


##### 1. Server

- `github.com/CloudyKit/jet/v6`: Template engine
- `github.com/bmizerany/pat`: URL pattern muxer
- `github.com/gorilla/websocket`: WebSocket protocol

##### 2. Client

- `Bootstrap`: component style
- `github.com/jaredreich/notie`: notification
- `github.com/joewalnes/reconnecting-websocket`: automatically reconnect WebSocket


### API

##### 1. localhost:8080/

<pre>
html 폴더에 있는 home.jet 파일을 렌더링한다
</pre>

##### 2. localhost:8080/ws

<pre>
html 서버와의 연결을 websocket 프로토콜로 업그레이드한다
</pre>

##### 3. localhost:8080/static/

<pre>
파일 서버 static 폴더에 있는 정적 파일을 요청한다
</pre>


### Websockets

##### 1. WsEndpoint

<pre>
1. 특정 클라이언트의 html 서버와의 연결을 websocket 프로토콜로 업그레이드한다
   
2. 업그레이드가 성공하면 연결 정보를 clients 맵에 연결정보:사용자이름 형태로 저장한다
   
3. 연결된 클라이언트에게 현재 연결되어 있는 다른 클라이언트들의 사용자이름과 연결이 성공했음을 JSON 형식으로 인코딩하여 전달한다
   
4. ListenForWS 함수를 고루틴으로 실행해 방금 업그레이드된 클라이언트로부터 웹소켓을 통해 전달되는 메시지를 받는다
   
5. 전달받은 메시지는 JSON에서 WsPayload 구조체로 디코딩되어 wsChan 채널에 푸시된다
</pre>

##### 2. ListenToWSChannel

<pre>
1. WsEndpoint 함수가 특정 클라이언트로부터 전달받은 메시지를 디코딩하여 채널에 푸시하면
   
2. main 함수에서 고루틴으로 실행되는 ListenToWSChannel 함수가 하나씩 꺼내서 처리한다
   
3. WsPayload 구조체로 디코딩된 메시지는 액션 타입에 따라 switch문에서 분기 처리된다
   
4. 액션 타입이 "username"인 경우: 메시지의 주인 클라이언트의 사용자이름을 변경하고 
   변경된 사용자 리스트를 가져와 다른 모든 클라이언트에게 브로드캐스팅한다

5. 액션 타입이 "left"인 경우: 메시지의 주인 클라이언트의 연결 정보를 삭제하고 
   변경된 사용자 리스트를 가져와 다른 모든 클라이언트에게 브로드캐스팅한다

6. 액션 타입이 "broadcast"인 경우: 메시지의 주인 클라이언트가 입력한 채팅 내용을 
   사용자이름: 메시지 형태로 다른 모든 클라이언트에게 브로드캐스팅한다
</pre>

### Client

##### 0. socket.onmessage
<pre>
웹소켓을 통해 서버로부터 클라이언트로 메시지가 전달되는 경우
액션 타입에 따른 분기 처리가 이루어진다

1. "list_users" 액션 타입
   사용자이름이 명시된 클라이언트들이 웹소켓에 연결되어 있다면
   해당 클라이언트들의 사용자이름을 unordered list로 화면에 렌더링

2. "broadcast" 액션 타입
   채팅 내역에 전달받은 메시지를 추가한다

3. "enter" 액션 타입
   웹소켓에 연결된 클라이언트들의 사용자이름을 리스트로 렌더링하고
   클라이언트가 서버에 성공적으로 연결되었음을 알림으로 짧게 표시한다

# 1번과 2번은 서버로부터 웹소켓에 연결된 모든 클라이언트에게 브로드캐스팅되는 방식

# 3번은 새롭게 연결을 시도한 단일 클라이언트에게 서버로부터 직접적으로 전달되는 방식
</pre>

##### 1. socket.onopen

<pre>
1. 서버로부터 웹소켓을 통해 새롭게 연결된 단일 클라이언트에게
연결되어 있는 클라이언트들의 사용자이름 리스트를 전달한다

2. 해당 클라이언트는 액션 타입 "enter"에 해당하는 메시지 분기 처리를 실행한다
</pre>

##### 2. userField onchange event handler

<pre>
사용자이름이 변경된 경우, 아래의 객체를 생성하고

{
    "action": "username",
    "username": this.value,
}

JSON 형식으로 인코딩하여 웹소켓을 통해 서버로 전달한다

서버는 이에 대한 응답으로 웹소켓에 연결되어 있는 모든 클라이언트의 사용자이름 리스트를 모든 클라이언트에게 브로드캐스팅한다
</pre>

##### 3. messageField keyup or sendBtn click event handler 

<pre>
클라이언트가 메시지 필드에 메시지를 입력하고 엔터를 누르거나 Send Message 버튼을 누른 경우, 

아래의 객체를 생성하고

{
    "action": "broadcast",
    "username": userField.value,
    "message": messageField.value,
}

JSON 형식으로 인코딩하여 웹소켓을 통해 서버로 전달한다

서버는 이에 대한 응답으로 웹소켓에 연결되어 있는 모든 클라이언트에게 입력된 메시지를 브로드캐스팅한다
</pre>

##### 4. window.onbeforeunload

<pre>
1. 사용자가 페이지를 떠날 때 클라이언트는 액션 타입을 "left"로 지정한 메시지를 서버에게 보냄으로써 
   사용자의 연결이 끊어진다는 것을 통지한다

2. 서버는 맵에 저장된 클라이언트의 연결 정보를 제거하고 갱신된 사용자이름 리스트를 다른 클라이언트들에게 브로드캐스팅한다
</pre>