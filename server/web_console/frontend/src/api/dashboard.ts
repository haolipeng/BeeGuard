import request from './request';
import { statusApi } from './system';

export const dashboardApi = {
  getHostStatusSummary: () => request.get('/views/host-status-summary'),
  getThreatTypeTotalCount: () => request.get('/views/threat-type-total-count'),
  getAlertHourlyStats: () => request.get('/views/alert-hourly-stats'),
  getAlertMonthlyStats: () => request.get('/views/alert-monthly-stats'),
  getSecurityAlertDailyStats: () => request.get('/views/security-alert-daily-stats'),
  getHostVulnTop5: () => request.get('/views/host-vuln-package-top5'),
  getHostBaselineFailTop5: () => request.get('/views/host-baseline-fail-top5'),
  getServiceStatusOverview: () => statusApi.getOverview(),
  getVulnChartData: () => request.get('/views/vuln-chart-data'),
  getImageVulnTop5: () => request.get('/views/image-vuln-top5'),
  getBaselineItemTop5: () => request.get('/views/baseline-item-top5'),
};
