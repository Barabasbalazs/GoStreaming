<template>
  <div>
    <button @click="createConnection">Connect to Server</button>
    <video :srcObject="stream" width="300" height="200" autoplay></video>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { OfferModel } from './models/offer-model';
import { getSd } from './utils/get-sd';

const stream = ref();

const pc = new RTCPeerConnection({
  iceServers: [{
    urls: 'stun:stun.l.google.com:19302'
  }]
});

pc.ontrack = (event) => {
  stream.value = event.streams[0];
}
pc.addTransceiver('video', {
  direction: 'sendrecv'
});
pc.createOffer().then(d => pc.setLocalDescription(d)).catch();

const createConnection = async () => {
  try {
    const resp = await getSd({ offer: btoa(JSON.stringify(pc.localDescription))});
    if (!resp) {
      console.log('Error while fetching');
      return;
    }
    // console.log(atob(resp.offer));
    pc.setRemoteDescription(JSON.parse(atob(resp.offer)));
  } catch (e) {
    console.log(e);
  }
}

</script>

<style scoped>
button {
  border: 2px;
  border-color: black;
}
</style>