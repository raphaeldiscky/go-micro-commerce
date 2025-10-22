import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { ChartConfig } from '@/components/ui/chart'
import {
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart'
import { mockRevenueChartData, mockRevenueStats } from '@/mocks/analytics'
import { ArrowDown, ArrowUp, DollarSign, TrendingUp } from 'lucide-react'
import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from 'recharts'

const chartConfig = {
  revenue: {
    color: 'hsl(var(--chart-1))',
    label: 'Revenue',
  },
} satisfies ChartConfig

export function RevenueOverview() {
  const { averageOrderValue, growthPercentage, totalRevenue } = mockRevenueStats

  const isPositiveGrowth = growthPercentage >= 0

  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Revenue</CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${totalRevenue.toLocaleString()}
            </div>
            <p className="text-xs text-muted-foreground">Last 30 days</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Growth Rate</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="flex items-center gap-2">
              <span className="text-2xl font-bold">
                {Math.abs(growthPercentage)}%
              </span>
              {isPositiveGrowth ? (
                <ArrowUp className="h-4 w-4 text-green-500" />
              ) : (
                <ArrowDown className="h-4 w-4 text-red-500" />
              )}
            </div>
            <p className="text-xs text-muted-foreground">vs previous month</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Avg Order Value
            </CardTitle>
            <DollarSign className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              ${averageOrderValue.toFixed(2)}
            </div>
            <p className="text-xs text-muted-foreground">Per transaction</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Revenue Trend</CardTitle>
        </CardHeader>
        <CardContent>
          <ChartContainer
            config={chartConfig}
            className="h-[250px] w-full md:h-[300px] lg:h-[350px]"
          >
            <AreaChart
              data={mockRevenueChartData}
              margin={{
                bottom: 0,
                left: 12,
                right: 12,
                top: 12,
              }}
            >
              <defs>
                <linearGradient id="fillRevenue" x1="0" y1="0" x2="0" y2="1">
                  <stop
                    offset="5%"
                    stopColor="var(--color-revenue)"
                    stopOpacity={0.8}
                  />
                  <stop
                    offset="95%"
                    stopColor="var(--color-revenue)"
                    stopOpacity={0.1}
                  />
                </linearGradient>
              </defs>
              <CartesianGrid vertical={false} />
              <XAxis
                dataKey="date"
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(value) => {
                  const date = new Date(value)
                  return `${date.getMonth() + 1}/${date.getDate()}`
                }}
              />
              <YAxis
                tickLine={false}
                axisLine={false}
                tickMargin={8}
                tickFormatter={(value) => `$${value}`}
              />
              <ChartTooltip
                cursor={false}
                content={
                  <ChartTooltipContent
                    indicator="line"
                    labelFormatter={(value) => {
                      const date = new Date(value)
                      return date.toLocaleDateString('en-US', {
                        day: 'numeric',
                        month: 'short',
                      })
                    }}
                    formatter={(value, name, item) => (
                      <>
                        <div
                          className="h-2.5 w-2.5 shrink-0 rounded-[2px] bg-[--color-bg]"
                          style={
                            {
                              '--color-bg': `var(--color-${name})`,
                            } as React.CSSProperties
                          }
                        />
                        {chartConfig[name as keyof typeof chartConfig].label ||
                          name}
                        <div className="ml-auto flex items-baseline gap-0.5 font-mono font-medium tabular-nums text-foreground">
                          ${value}
                          <span className="font-normal text-muted-foreground">
                            ({item.payload.orders} orders)
                          </span>
                        </div>
                      </>
                    )}
                  />
                }
              />
              <Area
                type="monotone"
                dataKey="revenue"
                stroke="var(--color-revenue)"
                strokeWidth={2}
                fill="url(#fillRevenue)"
              />
            </AreaChart>
          </ChartContainer>
        </CardContent>
      </Card>
    </div>
  )
}
