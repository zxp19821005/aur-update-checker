<template>
  <a-card title="关于 AUR更新检查器" :bordered="false">
    <a-descriptions :column="2" :bordered="true">
      <a-descriptions-item label="应用名称">AUR更新检查器</a-descriptions-item>
      <a-descriptions-item label="版本">1.0.0</a-descriptions-item>
      <a-descriptions-item label="描述">一个用于检查AUR软件包更新的工具</a-descriptions-item>
      <a-descriptions-item label="技术栈">
        <a-space>
          <a-tag color="blue">Vue 3</a-tag>
          <a-tag color="cyan">Go</a-tag>
          <a-tag color="green">Gin</a-tag>
          <a-tag color="red">Ant Design Vue</a-tag>
          <a-tag color="purple">GORM</a-tag>
          <a-tag color="orange">SQLite</a-tag>
        </a-space>
      </a-descriptions-item>
      <a-descriptions-item label="作者">Woodenman3993</a-descriptions-item>
      <a-descriptions-item label="版权">© {{ new Date().getFullYear() }} AUR更新检查器</a-descriptions-item>
    </a-descriptions>

    <a-divider />

    <a-alert
      message="开源许可"
      description="本项目采用 MIT 许可证开源，详情请参阅项目文档。"
      type="info"
      show-icon
    />

    <a-divider />

    <a-typography>
      <a-typography-title :level="4">功能特点</a-typography-title>
      <a-typography-paragraph>
        <ul>
          <li>添加、编辑、删除软件包</li>
          <li>自动从AUR获取当前版本和更新日期</li>
          <li>根据上游URL和版本提取关键字获取上游最新版本</li>
          <li>支持定时检查功能</li>
          <li>软件界面支持列表显示和卡片显示</li>
          <li>分级日志系统，支持彩色显示</li>
          <li>数据库存储软件包信息、AUR信息和上游信息</li>
          <li>支持检查测试版本功能</li>
          <li>多种上游检查器支持（GitHub、GitLab、Gitee、HTTP、JSON、NPM、PyPI等）</li>
        </ul>
      </a-typography-paragraph>

      <a-typography-title :level="4">使用说明</a-typography-title>
      <a-typography-paragraph>
        <h5>添加软件包</h5>
        <p>1. 点击"软件包管理"页面中的"添加软件包"按钮</p>
        <p>2. 填写软件包名称、上游URL、版本提取关键字和检查器类型</p>
        <p>3. 选择是否检查测试版本</p>
        <p>4. 点击"确定"保存</p>

        <h5>检查版本更新</h5>
        <p>1. 在"软件包管理"页面中，可以单独检查某个软件包的AUR版本或上游版本</p>
        <p>2. 也可以批量检查所有软件包的AUR版本或上游版本</p>
        <p>3. 检查结果会显示在软件包列表中</p>

        <h5>设置定时任务</h5>
        <p>1. 在"设置"页面中，可以设置定时检查任务</p>
        <p>2. 设置检查间隔（分钟）</p>
        <p>3. 点击"启动定时任务"按钮</p>

        <h5>查看日志</h5>
        <p>1. 在"日志"页面中，可以查看应用运行日志</p>
        <p>2. 可以按日志级别筛选</p>
        <p>3. 支持自动刷新功能</p>

        <h5>API 服务</h5>
        <p>1. 后端提供 RESTful API 服务</p>
        <p>2. 默认运行在 http://localhost:8080</p>
        <p>3. 支持跨域请求，便于前端开发</p>
        <p>4. 提供 Swagger 文档，方便接口调试</p>
      </a-typography-paragraph>
    </a-typography>
  </a-card>
</template>

<style scoped>
.ant-card {
  max-width: 1000px;
  margin: 0 auto;
}
</style>
