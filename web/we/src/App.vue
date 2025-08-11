<template>
  <div class="container">
    <div class="left">
      <div class="line-div">
        <el-input
          v-model="input1"
          placeholder="请输入升级程序的url链接地址"
        ></el-input>
        <el-button @click="handleCMD('update', input1)" class="line-btn"
          >全量升级
        </el-button>
      </div>
      <div class="line-div">
        <el-input
          v-model="input4"
          placeholder="请输入差量升级程序的url链接地址"
        ></el-input>
        <el-button @click="handleCMD('patch', input4)" class="line-btn"
          >差量升级
        </el-button>
      </div>
      <div class="line-div">
        <el-input v-model="input2" placeholder="请输入命令"></el-input>
        <el-button @click="handleCMD('cmd', input2)" class="line-btn" plain
          >执行命令
        </el-button>
      </div>
      <div class="line-div">
        <el-input v-model="input3" placeholder="执行self命令"></el-input>
        <el-button @click="handleCMD('self', input3)" class="line-btn" plain
          >执行self命令
        </el-button>
      </div>
      <div class="line-div">
        <el-button-group class="ml-4">
          <el-button type="primary" plain @click="toggleDark()">
            <span class="ml-2">{{ isDark ? 'Dark' : 'Light' }}</span>
          </el-button>
          <el-button @click="handleCMD('version', '')">获取版本号</el-button>
          <el-button type="primary" plain @click="handleCheckVersion"
            >检测版本
          </el-button>
          <el-button @click="handleCMD('sudo', '')">sudo</el-button>
          <el-button @click="handleCMD('get', '')">get</el-button>
          <el-button type="danger" plain @click="handleCMD('delete', '')"
            >delete
          </el-button>
          <el-button @click="handleCMD('restart', '')">restart</el-button>
          <el-button type="warning" plain @click="handleCMD('uninstall', '')"
            >uninstall
          </el-button>
        </el-button-group>
      </div>
      <div class="line-div-btn">
        <el-button-group class="ml-4">
          <el-button @click="handleReadLog">查看日志</el-button>
          <el-button @click="handleClearLog">清空日志</el-button>
          <el-button type="danger" plain @click="handleCMD('clear', '')"
            >删除缓存
          </el-button>
          <el-button @click="handleTest">测试按钮</el-button>
          <el-button @click="handleGithub">github</el-button>
        </el-button-group>
      </div>
      <div class="line-div-btn">
        <el-button-group class="ml-4">
          <el-button @click="handleCMD('panic', '')">panic</el-button>
          <el-button @click="handleCMD('null', '')">空指针</el-button>
        </el-button-group>
      </div>
    </div>
    <div class="right">
      <div ref="logContainer" class="log-container">
        <div v-for="(log, index) in logs" :key="index" class="log-item">
          <pre v-html="log"></pre>
        </div>
      </div>
    </div>
  </div>

  <!--  <UseDark v-slot="{ isDark, toggleDark }">-->
  <!--    <button @click="toggleDark()">Is Dark: {{ isDark }}</button>-->
  <!--  </UseDark>-->
  <UpgradeDialog ref="upgradeRef" @handle-upgrade="handleUpgrade" />
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { showLoading, syntaxHighlight } from './utils/utils.ts'
import UpgradeDialog from './components/UpgradeDialog.vue'
import { useDark, useToggle } from '@vueuse/core'
// import { isDark } from '../../.vitepress/theme/composables/dark'

const isDark = useDark()
const toggleDark = useToggle(isDark)
const logs = ref<string[]>([])
const logContainer = ref<HTMLDivElement | null>(null)

const upgradeRef = ref<InstanceType<typeof UpgradeDialog> | null>(null)
const input1 = ref<string>()
const input2 = ref<string>()
const input3 = ref<string>()
const input4 = ref<string>()

const addLog = (context: string): void => {
  const newLog = `${new Date().toLocaleString()}: ${context}`
  logs.value.unshift(newLog)
  // 滚动到顶部
  if (logContainer.value) {
    logContainer.value.scrollTop = 0
  }
}
const showUpgradeDialog = (
  patchUrl: string | undefined,
  fullUrl: string | undefined,
  releaseNotes: string | undefined,
) => {
  if (upgradeRef.value) {
    upgradeRef.value.openUpgradeDialog(patchUrl, fullUrl, releaseNotes)
  }
}

const handleCMD = (action: string | undefined, data: string | undefined) => {
  fetchRunApi(action, { data: data })
}
const handleGithub = () => {
  window.open('https://github.com/xxl6097/go-service')
}
const handleCheckVersion = () => {
  fetchRunApi('checkversion', {}, function (json: any) {
    if (json && json.code === 0 && json.data) {
      showUpgradeDialog(
        json.data.patchUrl,
        json.data.fullUrl,
        json.data.releaseNotes,
      )
    }
  })
}

const handleUpgrade = (binUrl: string) => {
  fetchRunApi('confirm-upgrade', { data: binUrl })
}

const handleReadLog = () => {
  const host = window.origin
  window.open(`${host}/log/`)
}
const handleTest = () => {
  addLog('这是一条测试数据')
}
const handleClearLog = () => {
  logs.value = []
}

const fetchRunApi = (
  action: string | undefined,
  data: any,
  callback: (json: object) => void = () => {},
) => {
  const body = {
    action: action,
    data: data,
  }
  console.log('body', body)
  const loading = showLoading('请求中...')
  console.log('fetchApi', body)
  fetch('../api/cmd', {
    credentials: 'include',
    method: 'POST',
    body: JSON.stringify(body),
  })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      console.log('fetch result:', json)
      if (callback) {
        callback(json)
      }
      if (json.code === 0) {
        if (typeof json.data === 'string') {
          addLog(json.data)
        } else {
          const rawJson = JSON.stringify(json.data, null, 2)
          const highlightedJSON = syntaxHighlight(rawJson)
          addLog(highlightedJSON)
        }
      } else {
        addLog(json.msg)
      }
    })
    .catch(() => {
      //showErrorTips('配置失败')
    })
    .finally(() => {
      loading.close()
    })
}

const source = ref<EventSource>()

function initSSE() {
  const ssurl = `${window.location.origin}/api/sse-stream`
  try {
    addLog(`开始连接SSE:${ssurl}`)
    const s = new EventSource(ssurl)
    source.value = s
    s.onmessage = (event) => {
      console.log('收到消息:', event.data)
      addLog(event.data)
    }
    s.onopen = (e) => {
      console.log('SSE连接已建立', s.readyState) // readyState=1表示连接正常
      addLog('连接成功 ' + e.currentTarget?.toString())
      console.log('sse connect sucessully..', e)
    }
    s.onerror = (e) => {
      source.value?.close()
      source.value = undefined
      addLog('连接错误:' + JSON.stringify(e))
      console.log('onerror received a message', e)
      setTimeout(function () {
        initSSE()
      }, 5000)
    }
  } catch (e) {
    console.log('sse init err', e)
    addLog(`连接SSE识别:${JSON.stringify(e)}`)
    setTimeout(function () {
      initSSE()
    }, 5000)
  }
}

const fetchData = () => {
  fetch('../api/version', { credentials: 'include', method: 'GET' })
    .then((res) => {
      return res.json()
    })
    .then((json) => {
      if (json) {
        document.title = `aatest ${json.appVersion}`
      }
    })
}

initSSE()
fetchData()
handleCMD('version', '')
</script>

<style>
.container {
  width: 100%;
  overflow: hidden; /* 清除浮动影响 */
}

.left {
  float: left; /* PC 端左右浮动 */
  width: 30%;
  padding: 10px;
  box-sizing: border-box;
}
.right {
  float: left; /* PC 端左右浮动 */
  width: 70%;
  padding: 10px;
  box-sizing: border-box;
}

/* 手机端媒体查询 */
@media (max-width: 768px) {
  .left,
  .right {
    width: 100%; /* 宽度占满 */
  }
}

.log-container {
  height: auto;
  max-height: 880px;
  overflow-y: auto;
  margin-left: 1px;
  border: 1px solid #a8a1a1; /* 统一设置边框：宽度 2px，实线，深灰色 */
  padding: 4px;
}

.log-item {
  margin-bottom: 5px;
}

.line-div {
  display: flex;
  margin-left: 5px;
  margin-right: 5px;
  margin-bottom: 8px;
}

.line-div-btn {
  display: flex;
  margin-left: 5px;
  margin-right: 5px;
  margin-bottom: 8px;
}

.line-btn {
  margin-left: 10px;
}
</style>
