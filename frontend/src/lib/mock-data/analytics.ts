export interface RevenueStats {
  totalRevenue: number
  previousMonthRevenue: number
  growthPercentage: number
  averageOrderValue: number
}

export interface OrderStats {
  totalOrders: number
  pendingOrders: number
  processingOrders: number
  completedOrders: number
  cancelledOrders: number
  todayOrders: number
}

export interface RevenueChartData {
  date: string
  revenue: number
  orders: number
}

export interface OrderStatusDistribution {
  name: string
  value: number
  color: string
  [key: string]: string | number
}

export const mockRevenueStats: RevenueStats = {
  averageOrderValue: 125.5,
  growthPercentage: 12.5,
  previousMonthRevenue: 45000,
  totalRevenue: 125430.5,
}

export const mockOrderStats: OrderStats = {
  cancelledOrders: 12,
  completedOrders: 342,
  pendingOrders: 25,
  processingOrders: 48,
  todayOrders: 18,
  totalOrders: 427,
}

export const mockRevenueChartData: Array<RevenueChartData> = [
  { date: '2024-01-01', orders: 15, revenue: 1850 },
  { date: '2024-01-02', orders: 23, revenue: 2890 },
  { date: '2024-01-03', orders: 19, revenue: 2380 },
  { date: '2024-01-04', orders: 27, revenue: 3390 },
  { date: '2024-01-05', orders: 21, revenue: 2630 },
  { date: '2024-01-06', orders: 29, revenue: 3640 },
  { date: '2024-01-07', orders: 18, revenue: 2250 },
  { date: '2024-01-08', orders: 31, revenue: 3890 },
  { date: '2024-01-09', orders: 25, revenue: 3130 },
  { date: '2024-01-10', orders: 22, revenue: 2760 },
  { date: '2024-01-11', orders: 26, revenue: 3260 },
  { date: '2024-01-12', orders: 20, revenue: 2500 },
  { date: '2024-01-13', orders: 28, revenue: 3510 },
  { date: '2024-01-14', orders: 24, revenue: 3010 },
  { date: '2024-01-15', orders: 30, revenue: 3760 },
  { date: '2024-01-16', orders: 17, revenue: 2130 },
  { date: '2024-01-17', orders: 33, revenue: 4140 },
  { date: '2024-01-18', orders: 26, revenue: 3260 },
  { date: '2024-01-19', orders: 21, revenue: 2630 },
  { date: '2024-01-20', orders: 29, revenue: 3640 },
  { date: '2024-01-21', orders: 23, revenue: 2880 },
  { date: '2024-01-22', orders: 32, revenue: 4010 },
  { date: '2024-01-23', orders: 27, revenue: 3390 },
  { date: '2024-01-24', orders: 25, revenue: 3130 },
  { date: '2024-01-25', orders: 34, revenue: 4260 },
  { date: '2024-01-26', orders: 22, revenue: 2760 },
  { date: '2024-01-27', orders: 28, revenue: 3510 },
  { date: '2024-01-28', orders: 30, revenue: 3760 },
  { date: '2024-01-29', orders: 26, revenue: 3260 },
  { date: '2024-01-30', orders: 35, revenue: 4390 },
]

export const mockOrderStatusDistribution: Array<OrderStatusDistribution> = [
  { color: '#22c55e', name: 'Completed', value: 342 },
  { color: '#3b82f6', name: 'Processing', value: 48 },
  { color: '#f59e0b', name: 'Pending', value: 25 },
  { color: '#ef4444', name: 'Cancelled', value: 12 },
]
