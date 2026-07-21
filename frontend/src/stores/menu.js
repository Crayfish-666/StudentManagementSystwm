import { defineStore } from 'pinia'
import { ref } from 'vue'
import { menuApi } from '@/api/sys'

const viewModules = import.meta.glob('../views/**/*.vue')

const DEFAULT_MENUS = [
  {
    code: 'dashboard', title: '工作台', icon: 'Odometer', path: '/dashboard',
    children: [
      { code: 'cmp-dashboard', title: '管理驾驶舱', path: '/dashboard', component: 'views/Dashboard.vue' },
      { code: 'cmp-ranking', title: '综合分排行', path: '/cmp/ranking', component: 'views/cmp/ScoreRanking.vue' }
    ]
  },
  {
    code: 'ty', title: '团员发展', icon: 'Flag', path: '/ty',
    children: [
      { code: 'ty-application', title: '入团申请', path: '/ty/application', component: 'views/ty/ApplicationList.vue' },
      { code: 'ty-approval', title: '审批中心', path: '/ty/approval', component: 'views/ty/ApprovalCenter.vue' },
      { code: 'ty-recommendation-meeting', title: '支部推优大会', path: '/ty/recommendation-meeting', component: 'views/ty/RecommendationMeetingList.vue' },
      { code: 'ty-cultivation', title: '培养记录管理', path: '/ty/cultivation', component: 'views/ty/CultivationView.vue' },
      { code: 'ty-development-object', title: '发展对象管理', path: '/ty/development-object', component: 'views/ty/DevelopmentObjectView.vue' },
      { code: 'ty-political-review', title: '政治审查管理', path: '/ty/political-review', component: 'views/ty/PoliticalReviewView.vue' },
      { code: 'ty-development-meeting', title: '接收发展大会', path: '/ty/development-meeting', component: 'views/ty/DevelopmentMeetingView.vue' },
      { code: 'ty-probationary', title: '预备团员转正', path: '/ty/probationary', component: 'views/ty/ProbationaryView.vue' },
      { code: 'ty-member-roster', title: '团员花名册', path: '/ty/member-roster', component: 'views/ty/MemberRoster.vue' }
    ]
  },
  {
    code: 'st', title: '社团活动', icon: 'Trophy', path: '/st',
    children: [
      { code: 'st-association', title: '社团管理', path: '/st/association', component: 'views/st/AssociationList.vue' },
      { code: 'st-recruit-plan', title: '招新计划管理', path: '/st/recruit-plan', component: 'views/st/RecruitPlanList.vue' },
      { code: 'st-recruit-apply', title: '招新申请广场', path: '/st/recruit-apply', component: 'views/st/RecruitApplyList.vue' },
      { code: 'st-activity', title: '活动管理与审批', path: '/st/activity', component: 'views/st/ActivityList.vue' }
    ]
  },
  {
    code: 'sq', title: '学生社区', icon: 'House', path: '/sq',
    children: [
      { code: 'sq-building', title: '楼栋与寝室网格', path: '/sq/building', component: 'views/sq/BuildingTree.vue' },
      { code: 'sq-inspection', title: '巡查记录大厅', path: '/sq/inspection', component: 'views/sq/InspectionList.vue' },
      { code: 'sq-incident', title: '异常事件处置', path: '/sq/incident', component: 'views/sq/IncidentList.vue' }
    ]
  },
  {
    code: 'qg', title: '勤工助学', icon: 'Briefcase', path: '/qg',
    children: [
      { code: 'qg-difficulty', title: '困难认定库', path: '/qg/difficulty', component: 'views/qg/DifficultyList.vue' },
      { code: 'qg-position', title: '岗位管理', path: '/qg/position', component: 'views/qg/PositionList.vue' },
      { code: 'qg-attendance', title: '工时打卡与考勤', path: '/qg/attendance', component: 'views/qg/AttendanceRecord.vue' }
    ]
  },
  {
    code: 'mine', title: '我的申请', icon: 'Document', path: '/mine',
    children: [
      { code: 'mine-ty-development', title: '我的团员发展', path: '/mine/ty-development', component: 'views/ty/MyDevelopment.vue' },
      { code: 'mine-ty-application', title: '我的入团申请', path: '/mine/ty-application', component: 'views/ty/ApplicationList.vue' },
      { code: 'mine-thought-report', title: '我的思想汇报', path: '/mine/thought-report', component: 'views/ty/MyThoughtReport.vue' },
      { code: 'mine-activity', title: '我的社团履历', path: '/mine/activity', component: 'views/st/ActivityList.vue' },
      { code: 'mine-work', title: '我的勤工记录', path: '/mine/work', component: 'views/qg/AttendanceRecord.vue' },
      { code: 'mine-score', title: '我的综合分', path: '/mine/score', component: 'views/cmp/MyScore.vue' },
      { code: 'mine-profile', title: '我的学籍档案', path: '/mine/profile', component: 'views/idx/MyProfile.vue' }
    ]
  },
  {
    code: 'idx', title: '学生管理', icon: 'User', path: '/idx',
    children: [
      { code: 'idx-student', title: '学生列表与履历', path: '/idx/student', component: 'views/idx/StudentList.vue' },
      { code: 'idx-import', title: '学生批量导入', path: '/idx/import', component: 'views/idx/StudentImport.vue' }
    ]
  },
  {
    code: 'sys', title: '系统管理', icon: 'Setting', path: '/sys',
    children: [
      { code: 'sys-dict', title: '字典管理', path: '/sys/dict', component: 'views/sys/DictManage.vue' },
      { code: 'sys-user', title: '用户账号管理', path: '/sys/user', component: 'views/sys/UserManage.vue' },
      { code: 'sys-org', title: '组织机构树', path: '/sys/org', component: 'views/sys/OrgManage.vue' },
      { code: 'sys-job', title: '任务监控', path: '/sys/job', component: 'views/sys/JobMonitor.vue' }
    ]
  }
]

export const useMenuStore = defineStore('menu', () => {
  const menuList = ref(DEFAULT_MENUS)
  const isLoaded = ref(false)

  async function fetchMenus(router) {
    try {
      const data = await menuApi.getMyMenus()
      if (data && data.menus && data.menus.length) {
        menuList.value = data.menus
      } else {
        menuList.value = DEFAULT_MENUS
      }
      isLoaded.value = true

      if (router) {
        registerDynamicRoutes(router)
      }
    } catch (err) {
      console.warn('获取服务端动态菜单失败，降级使用内建 35 视图菜单树:', err)
      menuList.value = DEFAULT_MENUS
      isLoaded.value = true
      if (router) {
        registerDynamicRoutes(router)
      }
    }
  }

  function registerDynamicRoutes(router) {
    const layoutRoute = router.getRoutes().find(r => r.name === 'Layout')
    if (!layoutRoute) return

    const dynamicRoutes = buildRoutes(menuList.value)
    dynamicRoutes.forEach(route => {
      if (!router.hasRoute(route.name)) {
        router.addRoute('Layout', route)
      }
    })
  }

  function buildRoutes(menus) {
    const routes = []
    for (const menu of menus) {
      if (menu.children && menu.children.length) {
        routes.push(...buildRoutes(menu.children))
      } else if (menu.component) {
        const moduleKey = `../${menu.component}`
        const componentLoader = viewModules[moduleKey]
        if (componentLoader) {
          routes.push({
            name: menu.code,
            path: menu.path,
            component: componentLoader,
            meta: { title: menu.title, requiresAuth: true }
          })
        }
      }
    }
    return routes
  }

  function resetMenus() {
    menuList.value = DEFAULT_MENUS
    isLoaded.value = false
  }

  return {
    menuList,
    isLoaded,
    fetchMenus,
    resetMenus
  }
})
