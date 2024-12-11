import Vue from "vue";
import VueRouter from "vue-router";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    name: "Home",
    redirect: { name: "Dashboard" },
  },
  {
    path: "/dashboard",
    name: "Dashboard",
    component: () => import("../views/Dashboard.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/namespaces",
    name: "Namespaces",
    component: () => import("../views/Namespaces.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/namespaces/:namespaceId",
    name: "NamespacePage",
    component: () => import("../views/NamespacePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/accounts",
    name: "Accounts",
    component: () => import("../views/Accounts.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/accounts/:accountId",
    name: "Account",
    component: () => import("../views/AccountPage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sp",
    name: "ServicesProviders",
    component: () => import("../views/ServicesProviders.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sp/create",
    name: "ServicesProviders create",
    component: () => import("../views/ServicesProvidersCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sp/showcases",
    name: "ServicesProviders showcases",
    component: () => import("../views/Showcases.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sp/:uuid",
    name: "ServicesProvider",
    component: () => import("../views/ServicesProvidersPage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sp/:uuid/edit",
    name: "ServicesProvider edit",
    component: () => import("../views/ServicesProvidersCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/dns",
    name: "DNS manager",
    component: () => import("../views/dnsManager.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/dns/:dnsname",
    name: "Zone manager",
    component: () => import("../views/ZoneManager.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/chats",
    name: "Chats",
    redirect: {
      name: "Plugin",
      params: { title: "Chats", url: "/cc.ui/" },
      query: {
        url: "/cc.ui/",
        fullscreen: window.innerWidth > 768 ? "false" : "true",
      },
    },
  },
  {
    path: "/chat/:uuid",
    name: "Chat",
    redirect: (to) => {
      return {
        name: "Plugin",
        params: {
          title: "Chats",
          url: "/cc.ui/",
          params: {
            redirect: `dashboard/${to.params.uuid}`,
          },
        },
        query: { url: "/cc.ui/" },
      };
    },
  },
  {
    path: "/settings",
    name: "Settings",
    component: () => import("../views/Settings.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/settings/app",
    name: "AppSetting",
    component: () => import("../views/AppSettings.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/settings/widget",
    name: "WidgetSetting",
    component: () => import("../views/WidgetSettings.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/settings/plugins",
    name: "PluginsSetting",
    component: () => import("../views/PluginsSettings.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/settings/invoices",
    name: "InvoicesSetting",
    component: () => import("../views/InvoicesSettings.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/plugin/:title",
    name: "Plugin",
    component: () => import("../views/PluginPage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/services",
    name: "Services",
    component: () => import("../views/Services.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/services/create",
    name: "Service create",
    component: () => import("../views/ServiceCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/services/:serviceId",
    name: "Service",
    component: () => import("../views/ServicePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/services/:serviceId/edit",
    name: "Service edit",
    component: () => import("../views/ServiceCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances",
    name: "Instances",
    component: () => import("../views/Instances.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances/create",
    name: "Instance create",
    component: () => import("../views/InstanceCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances/:instanceId/edit/",
    name: "Instance edit",
    component: () => import("../views/InstanceCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances/:instanceId",
    name: "Instance",
    component: () => import("../views/InstancePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances/:instanceId/vnc",
    component: () => import("../views/Vnc.vue"),
    name: "Vnc",
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/instances/:instanceId/dns",
    component: () => import("../views/InstanceDNS.vue"),
    name: "InstanceDns",
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/login",
    name: "Login",
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/login.vue"),
    meta: {
      requireUnlogin: true,
    },
  },
  {
    path: "/plans",
    name: "Plans",
    component: () => import("../views/Plans.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/plans/create",
    name: "Plans create",
    component: () => import("../views/PlansCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/plans/:planId",
    name: "Plan",
    component: () => import("../views/PlanPage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/transactions",
    name: "Transactions",
    component: () => import("../views/Transactions.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/reports",
    name: "Reports",
    component: () => import("../views/Reports.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/transactions/create",
    name: "Transactions create",
    component: () => import("../views/TransactionsCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/transactions/:uuid",
    name: "Transaction edit",
    component: () => import("../views/TransactionsCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/currencies",
    name: "Currencies",
    component: () => import("../views/Currencies.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/sessions",
    name: "Sessions",
    component: () => import("../views/Sessions.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/history",
    name: "History",
    component: () => import("../views/History.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/showcases",
    name: "Showcases",
    component: () => import("../views/Showcases.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/showcases/:uuid",
    name: "ShowcasePage",
    component: () => import("../views/ShowcasePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/showcases/create",
    name: "CreateShowcase",
    component: () => import("../views/ShowcaseCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/addons",
    name: "Addons",
    component: () => import("../views/Addons.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/addons/create",
    name: "Addon create",
    component: () => import("../views/AddonCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/addons/:uuid",
    name: "Addon page",
    component: () => import("../views/AddonPage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/invoices",
    name: "Invoices",
    component: () => import("../views/Invoices.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/invoices/create",
    name: "Invoice create",
    component: () => import("../views/InvoiceCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/invoices/:uuid",
    name: "Invoice page",
    component: () => import("../views/InvoicePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/promocodes",
    name: "Promocodes",
    component: () => import("../views/Promocodes.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/promocodes/create",
    name: "Promocode create",
    component: () => import("../views/PromocodeCreate.vue"),
    meta: {
      requireLogin: true,
    },
  },
  {
    path: "/promocodes/:uuid",
    name: "Promocode page",
    component: () => import("../views/PromocodePage.vue"),
    meta: {
      requireLogin: true,
    },
  },
];

const router = new VueRouter({
  mode: "history",
  base: process.env.BASE_URL,
  routes,
});

export default router;
