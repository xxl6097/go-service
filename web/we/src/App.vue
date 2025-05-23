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
          <el-button @click="handleCMD(input2, '')" style="width: 50%" plain
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
          <el-button @click="handleCMD('version', '')">获取版本号</el-button>
          <el-button @click="handleCMD('sudo', '')">sudo</el-button>
          <el-button @click="handleCMD('get', '')">get</el-button>
          <el-button @click="handleCMD('delete', '')">delete</el-button>
          <el-button @click="handleCMD('restart', '')">restart</el-button>
          <el-button @click="handleCMD('uninstall', '')">uninstall</el-button>
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
          <el-button @click="handleCMD('clear', '')">清空数据</el-button>
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
