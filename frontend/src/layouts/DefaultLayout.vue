<template>
  <el-container class="sh-layout-container">
    <!-- Top Glassmorphism Header -->
    <el-header class="sh-layout-header" height="68px">
      <!-- Left: Logo & App Title -->
      <div class="header-left" @click="$router.push('/dashboard')">
        <div class="logo-icon-wrap">
          <el-icon :size="24"><School /></el-icon>
        </div>
        <div class="brand-titles">
          <span class="main-title">StudentHub</span>
          <span class="sub-title">学生一站式自主管理系统</span>
        </div>
      </div>

      <!-- Right: Actions & User Info -->
      <div class="header-right">
        <!-- Role Switcher Dropdown -->
        <div class="role-switcher-box">
          <el-icon><Switch /></el-icon>
          <el-select v-model="currentRole" placeholder="切换角色视角" size="default" style="width: 140px;">
            <el-option label="系统管理员" value="R-SY-ADMIN" />
            <el-option label="院系辅导员" value="R-COL-COUN" />
            <el-option label="院系团委书记" value="R-COL-LEAGUE" />
            <el-option label="社团社长/干部" value="R-STU-ASSOC" />
            <el-option label="普通学生" value="R-STU-NORM" />
          </el-select>
        </div>

        <!-- Notification Bell -->
        <div class="noti-bell-wrap" @click="$router.push('/notifications')">
          <el-badge :value="3" class="noti-badge">
            <el-icon :size="20"><Bell /></el-icon>
          </el-badge>
        </div>

        <div class="divider"></div>

        <!-- User Profile Dropdown -->
        <el-dropdown trigger="click" @command="handleUserCommand">
          <div class="user-profile-trigger">
            <el-avatar :size="36" icon="UserFilled" class="user-avatar" />
            <span class="user-name">{{ displayName }}</span>
            <el-icon><ArrowDown /></el-icon>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">
                <el-icon><User /></el-icon>
                <span>我的档案</span>
              </el-dropdown-item>
              <el-dropdown-item command="password">
                <el-icon><Key /></el-icon>
                <span>修改密码</span>
              </el-dropdown-item>
              <el-dropdown-item divided command="logout">
                <el-icon><SwitchButton /></el-icon>
                <span>退出登录</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <!-- Main Layout Body -->
    <el-container class="sh-layout-body">
      <!-- Left Sidebar Menu -->
      <el-aside :width="isCollapse ? '68px' : '230px'" class="sh-layout-aside">
        <div class="sidebar-toggle" @click="isCollapse = !isCollapse">
          <el-icon :size="16">
            <component :is="isCollapse ? 'Expand' : 'Fold'" />
          </el-icon>
        </div>

        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapse"
          :collapse-transition="false"
          router
          class="sh-sidebar-menu"
        >
          <!-- 1. 工作台 -->
          <el-sub-menu index="/dashboard">
            <template #title>
              <el-icon><Odometer /></el-icon>
              <span>工作台</span>
            </template>
            <el-menu-item index="/dashboard">管理驾驶舱</el-menu-item>
            <el-menu-item index="/cmp/ranking">综合分排行</el-menu-item>
          </el-sub-menu>

          <!-- 2. 团员发展 -->
          <el-sub-menu index="/ty">
            <template #title>
              <el-icon><Flag /></el-icon>
              <span>团员发展</span>
            </template>
            <el-menu-item index="/ty/application">入团申请</el-menu-item>
            <el-menu-item index="/ty/approval">审批中心</el-menu-item>
            <el-menu-item index="/ty/recommendation-meeting">推优大会</el-menu-item>
            <el-menu-item index="/ty/cultivation">培养记录</el-menu-item>
            <el-menu-item index="/ty/development-object">发展对象</el-menu-item>
            <el-menu-item index="/ty/political-review">政审管理</el-menu-item>
            <el-menu-item index="/ty/development-meeting">发展大会</el-menu-item>
            <el-menu-item index="/ty/probationary">转正流程</el-menu-item>
            <el-menu-item index="/ty/member-roster">团员花名册</el-menu-item>
          </el-sub-menu>

          <!-- 3. 社团活动 -->
          <el-sub-menu index="/st">
            <template #title>
              <el-icon><Trophy /></el-icon>
              <span>社团活动</span>
            </template>
            <el-menu-item index="/st/association">社团管理</el-menu-item>
            <el-menu-item index="/st/recruit-plan">招新计划</el-menu-item>
            <el-menu-item index="/st/recruit-apply">招新广场</el-menu-item>
            <el-menu-item index="/st/activity">活动管理</el-menu-item>
          </el-sub-menu>

          <!-- 4. 学生社区 -->
          <el-sub-menu index="/sq">
            <template #title>
              <el-icon><House /></el-icon>
              <span>学生社区</span>
            </template>
            <el-menu-item index="/sq/building">楼栋网格</el-menu-item>
            <el-menu-item index="/sq/inspection">巡查记录</el-menu-item>
            <el-menu-item index="/sq/incident">异常处置</el-menu-item>
          </el-sub-menu>

          <!-- 5. 勤工助学 -->
          <el-sub-menu index="/qg">
            <template #title>
              <el-icon><Briefcase /></el-icon>
              <span>勤工助学</span>
            </template>
            <el-menu-item index="/qg/difficulty">困难认定</el-menu-item>
            <el-menu-item index="/qg/position">岗位管理</el-menu-item>
            <el-menu-item index="/qg/attendance">工时打卡</el-menu-item>
          </el-sub-menu>

          <!-- 6. 我的申请 -->
          <el-sub-menu index="/mine">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>我的申请</span>
            </template>
            <el-menu-item index="/mine/ty-development">我的团员发展</el-menu-item>
            <el-menu-item index="/mine/ty-application">我的入团申请</el-menu-item>
            <el-menu-item index="/mine/thought-report">我的思想汇报</el-menu-item>
            <el-menu-item index="/mine/activity">我的社团</el-menu-item>
            <el-menu-item index="/mine/work">我的勤工</el-menu-item>
            <el-menu-item index="/mine/score">我的综合分</el-menu-item>
            <el-menu-item index="/mine/profile">我的档案</el-menu-item>
          </el-sub-menu>

          <!-- 7. 学生管理 -->
          <el-sub-menu index="/idx">
            <template #title>
              <el-icon><User /></el-icon>
              <span>学生管理</span>
            </template>
            <el-menu-item index="/idx/student">学生列表</el-menu-item>
            <el-menu-item index="/idx/import">学生导入</el-menu-item>
          </el-sub-menu>

          <!-- 8. 系统管理 -->
          <el-sub-menu index="/sys">
            <template #title>
              <el-icon><Setting /></el-icon>
              <span>系统管理</span>
            </template>
            <el-menu-item index="/sys/dict">字典管理</el-menu-item>
            <el-menu-item index="/sys/user">用户管理</el-menu-item>
            <el-menu-item index="/sys/org">组织管理</el-menu-item>
            <el-menu-item index="/sys/job">任务监控</el-menu-item>
          </el-sub-menu>
        </el-menu>
      </el-aside>

      <!-- Right Content Container -->
      <el-main class="sh-layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  School, Switch, Bell, ArrowDown, User, Key, SwitchButton,
  Expand, Fold, Odometer, Flag, Trophy, House, Briefcase, Document, Setting
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isCollapse = ref(false)
const currentRole = ref('R-SY-ADMIN')
const activeMenu = computed(() => route.path)
const displayName = computed(() => authStore.displayName || '管理员')

async function handleUserCommand(cmd) {
  if (cmd === 'logout') {
    await authStore.logout()
    ElMessage.info('已安全退出登录')
    router.push('/login')
  } else if (cmd === 'profile') {
    router.push('/mine/profile')
  } else if (cmd === 'password') {
    ElMessage.info('功能提示：请在个人中心修改密码')
  }
}
</script>

<style scoped>
.sh-layout-container {
  min-height: 100vh;
  background: var(--sh-bg-main);
}

/* Header */
.sh-layout-header {
  background: var(--sh-bg-header);
  backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--sh-border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 24px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
}
.logo-icon-wrap {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  background: var(--sh-gradient-brand);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  box-shadow: var(--sh-shadow-glow);
}
.brand-titles {
  display: flex;
  flex-direction: column;
}
.main-title {
  font-family: 'Outfit', sans-serif;
  font-size: 18px;
  font-weight: 700;
  color: #ffffff;
}
.sub-title {
  font-size: 11px;
  color: var(--sh-text-secondary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}
.role-switcher-box {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--sh-text-secondary);
}
.noti-bell-wrap {
  cursor: pointer;
  padding: 8px;
  color: var(--sh-text-secondary);
  transition: color 0.2s;
}
.noti-bell-wrap:hover {
  color: var(--sh-primary);
}
.divider {
  width: 1px;
  height: 24px;
  background: var(--sh-border-color);
}
.user-profile-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}
.user-avatar {
  background: var(--sh-gradient-brand);
}
.user-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--sh-text-primary);
}

/* Sidebar */
.sh-layout-body {
  height: calc(100vh - 68px);
}
.sh-layout-aside {
  background: rgba(15, 23, 42, 0.6);
  border-right: 1px solid var(--sh-border-color);
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
}

.sidebar-toggle {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--sh-text-muted);
  cursor: pointer;
  border-bottom: 1px solid var(--sh-border-color);
  transition: color 0.2s;
}
.sidebar-toggle:hover {
  color: var(--sh-primary);
}

.sh-sidebar-menu {
  border: none;
  background: transparent;
  flex: 1;
  overflow-y: auto;
}

:deep(.el-sub-menu__title), :deep(.el-menu-item) {
  color: var(--sh-text-secondary) !important;
  font-weight: 500;
}
:deep(.el-sub-menu__title:hover), :deep(.el-menu-item:hover) {
  color: #ffffff !important;
  background: rgba(99, 102, 241, 0.12) !important;
}
:deep(.el-menu-item.is-active) {
  color: #ffffff !important;
  background: var(--sh-gradient-brand) !important;
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.35);
  border-radius: 8px;
  margin: 4px 8px;
}

/* Main */
.sh-layout-main {
  padding: 24px;
  overflow-y: auto;
}
</style>
