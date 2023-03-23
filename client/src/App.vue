<template>
  <div class="flex flex-col items-center space-y-2">
    <h1>Video Area</h1>
    <video :srcObject="stream" width="300" height="200" autoplay></video>
    <button class="border border-black p-2 rounded" @click="playVideo">
      Stream Video
    </button>
  </div>
  <div class="flex flex-col items-center space-y-2 mt-4">
    <h1>Messsage Area</h1>
    <div class="flex flex-col border border-black rounded w-1/3 h-full">
      <p>{{ startDate }}</p>
      <p v-for="text in chatText">{{ text }}</p>
    </div>
    <div class="border border-black rounded w-1/3 h-full">
      <textarea class="w-full h-full"
        v-model=messageText
      ></textarea>
    </div>
    <button class="border border-black p-2 rounded" @click="sendMessage">Send</button>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getSd } from "./utils/get-sd";

const stream = ref();
const pc = ref();
const wsConn = ref();
const chatText = ref(['']);
const messageText = ref('');
const startDate = new Date().toISOString();

const setupRTC = async () => {
  pc.value = new RTCPeerConnection({
      iceServers: [
        {
          urls: "stun:stun.l.google.com:19302",
        },
      ],
    });
    pc.value.addTransceiver("video", {
      direction: "sendrecv",
    });
    pc.value.ontrack = (event: { streams: any[]; }) => {
      stream.value = event.streams[0];
    };

    const d = await pc.value.createOffer();
    pc.value.setLocalDescription(d);
}

const setupWS = () => {
  wsConn.value = new WebSocket("ws://localhost:8080/ws");
  wsConn.value.onmessage = (event: { data: string; }) => {
    chatText.value.push(event.data);
  }
}

onMounted(async () => {
  try {
    await setupRTC();
    setupWS();
  } catch (e) {
    console.log(e);
  }
});

const playVideo = async () => {
  try {
    const resp = await getSd({
      offer: btoa(JSON.stringify(pc.value.localDescription)),
    });
    if (!resp) {
      console.log("Error while fetching");
      return;
    }
    console.log(atob(resp.offer));
    pc.value.setRemoteDescription(JSON.parse(atob(resp.offer)));
  } catch (e) {
    console.log(e);
  }
};

const sendMessage = () => {
  wsConn.value.send(messageText.value);
  messageText.value = '';
}

</script>