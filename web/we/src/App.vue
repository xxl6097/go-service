<template>
  <div>
    <el-row style="margin-top: 10px">
      <el-col :span="10">
        <div style="display: flex; margin-left: 5px; margin-right: 5px">
          <el-input
            v-model="input1"
            placeholder="请输入升级程序的url链接地址"
            style="width: 50%"
          ></el-input>
          <el-button @click="handleCMD('update', input1)" style="width: 50%"
            >升级程序
          </el-button>
        </div>
        <div style="display: flex; margin-left: 5px; margin-right: 5px">
          <el-input
            v-model="input4"
            placeholder="请输入差量升级程序的url链接地址"
            style="width: 50%"
          ></el-input>
          <el-button @click="handleCMD('patch', input4)" style="width: 50%"
            >差量程序
          </el-button>
        </div>
        <div
          style="
            display: flex;
            margin-left: 5px;
            margin-right: 5px;
            margin-top: 10px;
          "
        >
          <el-input
            v-model="input2"
            style="width: 50%"
            placeholder="请输入命令"
          ></el-input>
          <el-button @click="handleCMD('cmd', input2)" style="width: 50%" plain
            >执行命令
          </el-button>
        </div>
        <div
          style="
            display: flex;
            margin-left: 5px;
            margin-right: 5px;
            margin-top: 10px;
          "
        >
          <el-input
            v-model="input3"
            style="width: 50%"
            placeholder="执行self命令"
          ></el-input>
          <el-button @click="handleCMD('self', input3)" style="width: 50%" plain
            >执行self命令
          </el-button>
        </div>
        <div
          style="
            display: flex;
            margin-left: 5px;
            margin-right: 5px;
            margin-top: 10px;
          "
        >
          <el-button @click="handleCMD('version', '')">获取版本号</el-button>
          <el-button type="primary" plain @click="handleCMD('checkversion', '')"
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
        </div>
        <div
          style="
            display: flex;
            margin-left: 5px;
            margin-right: 5px;
            margin-top: 10px;
          "
        >
          <el-button @click="handleReadLog">查看日志</el-button>
          <el-button @click="handleClearLog">清空日志</el-button>
          <el-button type="danger" plain @click="handleCMD('clear', '')"
            >删除缓存
          </el-button>
          <el-button @click="handleTest">测试按钮</el-button>
        </div>
      </el-col>
      <el-col :span="14">
        <el-card title="日志面板" class="log-container">
          <div>
            <div ref="logContainer" class="log-container">
              <div v-for="(log, index) in logs" :key="index" class="log-item">
                <pre v-html="log"></pre>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue'
import { showLoading, syntaxHighlight } from './utils/utils.ts'

const logs = ref<string[]>([])
const logContainer = ref<HTMLDivElement | null>(null)

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

const handleCMD = (action: string | undefined, data: string | undefined) => {
  fetchRunApi(action, { data: data })
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

const fetchRunApi = (action: string | undefined, data: any) => {
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

initSSE()
handleCMD('version', '')
</script>

<style>
#head {
  margin-bottom: 30px;
}

.log-container {
  height: auto;
  max-height: 800px;
  overflow-y: auto;
  margin-left: 20px;
}

.log-item {
  margin-bottom: 5px;
}
</style>
