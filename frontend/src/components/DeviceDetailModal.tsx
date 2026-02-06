import { useState, useEffect } from 'react'
import { apiClient } from '../api/client'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Area,
  AreaChart,
} from 'recharts'

interface DeviceDetailModalProps {
  device: any
  onClose: () => void
}

type Period = '1h' | '6h' | '24h' | '7d' | '30d'
type ChartType = 'response' | 'uptime'

const periodLabels: Record<Period, string> = {
  '1h': '1 Hour',
  '6h': '6 Hours',
  '24h': '24 Hours',
  '7d': '7 Days',
  '30d': '30 Days',
}

const periodDurations: Record<Period, number> = {
  '1h': 60 * 60 * 1000,
  '6h': 6 * 60 * 60 * 1000,
  '24h': 24 * 60 * 60 * 1000,
  '7d': 7 * 24 * 60 * 60 * 1000,
  '30d': 30 * 24 * 60 * 60 * 1000,
}

export default function DeviceDetailModal({ device, onClose }: DeviceDetailModalProps) {
  const [status, setStatus] = useState<any>(null)
  const [history, setHistory] = useState<any[]>([])
  const [errors, setErrors] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [period, setPeriod] = useState<Period>('24h')
  const [chartType, setChartType] = useState<ChartType>('response')

  useEffect(() => {
    loadDeviceData()
    const interval = setInterval(loadDeviceData, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [device.id, period])

  const loadDeviceData = async () => {
    try {
      const now = new Date()
      const start = new Date(now.getTime() - periodDurations[period])

      const [statusData, historyData, errorsData] = await Promise.all([
        apiClient.getDeviceStatus(device.id),
        apiClient.getDeviceHistory(
          device.id,
          start.toISOString(),
          now.toISOString()
        ),
        apiClient.getDeviceErrors(device.id, 10),
      ])
      setStatus(statusData)
      setHistory(historyData || [])
      setErrors(errorsData || [])
    } catch (error) {
      console.error('Failed to load device data:', error)
    } finally {
      setLoading(false)
    }
  }

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp * 1000)
    switch (period) {
      case '1h':
      case '6h':
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      case '24h':
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      case '7d':
        return date.toLocaleDateString([], { weekday: 'short' })
      case '30d':
        return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
      default:
        return date.toLocaleString()
    }
  }

  const formatResponseTime = (value: number) => {
    if (value === 0) return '0ms'
    if (value < 1) return `${(value * 1000).toFixed(0)}μs`
    return `${value.toFixed(1)}ms`
  }

  const chartData = history
    .map((point) => ({
      ...point,
      time: formatTime(point.timestamp),
      responseTime: point.response_time,
      uptime: point.status === 'online' ? 100 : 0,
    }))
    .sort((a, b) => a.timestamp - b.timestamp)

  // Calculate uptime percentage for the period
  const uptimePercentage = history.length > 0
    ? (history.filter(h => h.status === 'online').length / history.length * 100).toFixed(1)
    : '0.0'

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">{device.name}</h2>
            <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">{device.hostname}</p>
            {device.is_critical && (
              <span className="inline-block mt-2 text-xs bg-red-600 text-white px-2 py-1 rounded-full">
                CRITICAL DEVICE
              </span>
            )}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 text-2xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {loading ? (
            <div className="text-center py-12">
              <div className="text-lg text-gray-600 dark:text-gray-400">Loading...</div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Status Cards */}
              <div className="grid grid-cols-4 gap-4">
                <div className={`rounded-lg p-4 ${status?.status === 'online' ? 'bg-green-100 dark:bg-green-900/30' : 'bg-red-100 dark:bg-red-900/30'}`}>
                  <div className={`text-2xl font-bold ${status?.status === 'online' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'}`}>
                    {status?.status === 'online' ? 'ONLINE' : 'OFFLINE'}
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">Current Status</div>
                </div>
                <div className="bg-blue-100 dark:bg-blue-900/30 rounded-lg p-4">
                  <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                    {status?.response_time ? formatResponseTime(status.response_time) : 'N/A'}
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">Response Time</div>
                </div>
                <div className="bg-purple-100 dark:bg-purple-900/30 rounded-lg p-4">
                  <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
                    {uptimePercentage}%
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">Uptime</div>
                </div>
                <div className="bg-orange-100 dark:bg-orange-900/30 rounded-lg p-4">
                  <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">
                    {device.check_interval}s
                  </div>
                  <div className="text-sm text-gray-600 dark:text-gray-400">Check Interval</div>
                </div>
              </div>

              {/* Chart */}
              <div className="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                <div className="flex items-center justify-between mb-4">
                  <div className="flex gap-2">
                    <button
                      onClick={() => setChartType('response')}
                      className={`px-3 py-1 text-sm rounded ${
                        chartType === 'response'
                          ? 'bg-blue-600 text-white'
                          : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                      }`}
                    >
                      Response Time
                    </button>
                    <button
                      onClick={() => setChartType('uptime')}
                      className={`px-3 py-1 text-sm rounded ${
                        chartType === 'uptime'
                          ? 'bg-green-600 text-white'
                          : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-300 dark:hover:bg-gray-600'
                      }`}
                    >
                      Uptime
                    </button>
                  </div>
                  <div className="flex gap-1">
                    {(Object.keys(periodLabels) as Period[]).map((p) => (
                      <button
                        key={p}
                        onClick={() => setPeriod(p)}
                        className={`px-3 py-1 text-xs rounded ${
                          period === p
                            ? 'bg-blue-600 text-white'
                            : 'bg-gray-200 dark:bg-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-300 dark:hover:bg-gray-600'
                        }`}
                      >
                        {periodLabels[p]}
                      </button>
                    ))}
                  </div>
                </div>

                {chartData.length === 0 ? (
                  <div className="h-48 flex items-center justify-center text-gray-500 dark:text-gray-400">
                    No data available for this period
                  </div>
                ) : chartType === 'response' ? (
                  <ResponsiveContainer width="100%" height={250}>
                    <AreaChart data={chartData}>
                      <defs>
                        <linearGradient id="responseTimeGradient" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                          <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                        </linearGradient>
                      </defs>
                      <XAxis
                        dataKey="time"
                        axisLine={false}
                        tickLine={false}
                        tick={{ fill: '#9ca3af', fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        axisLine={false}
                        tickLine={false}
                        tick={{ fill: '#9ca3af', fontSize: 12 }}
                        tickFormatter={formatResponseTime}
                        width={60}
                      />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: '#1f2937',
                          border: '1px solid #374151',
                          borderRadius: '8px',
                          color: '#fff',
                        }}
                        labelStyle={{ color: '#9ca3af' }}
                        formatter={(value: number) => [formatResponseTime(value), 'Response Time']}
                      />
                      <Area
                        type="monotone"
                        dataKey="responseTime"
                        stroke="#3b82f6"
                        strokeWidth={2}
                        fill="url(#responseTimeGradient)"
                      />
                    </AreaChart>
                  </ResponsiveContainer>
                ) : (
                  <ResponsiveContainer width="100%" height={250}>
                    <LineChart data={chartData}>
                      <XAxis
                        dataKey="time"
                        axisLine={false}
                        tickLine={false}
                        tick={{ fill: '#9ca3af', fontSize: 12 }}
                        interval="preserveStartEnd"
                      />
                      <YAxis
                        domain={[0, 100]}
                        axisLine={false}
                        tickLine={false}
                        tick={{ fill: '#9ca3af', fontSize: 12 }}
                        tickFormatter={(v) => `${v}%`}
                        width={50}
                      />
                      <Tooltip
                        contentStyle={{
                          backgroundColor: '#1f2937',
                          border: '1px solid #374151',
                          borderRadius: '8px',
                          color: '#fff',
                        }}
                        labelStyle={{ color: '#9ca3af' }}
                        formatter={(value: number) => [`${value === 100 ? 'Online' : 'Offline'}`, 'Status']}
                      />
                      <Line
                        type="stepAfter"
                        dataKey="uptime"
                        stroke="#22c55e"
                        strokeWidth={2}
                        dot={false}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                )}
              </div>

              {/* Recent Errors */}
              {errors.length > 0 && (
                <div className="bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg p-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
                    Recent Errors (Last 10)
                  </h3>
                  <div className="space-y-2">
                    {errors.map((error, idx) => (
                      <div
                        key={idx}
                        className="flex items-start gap-3 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-900 rounded-lg"
                      >
                        <div className="flex-shrink-0 w-2 h-2 mt-2 bg-red-500 rounded-full"></div>
                        <div className="flex-1 min-w-0">
                          <div className="text-sm text-gray-600 dark:text-gray-400">
                            {new Date(error.timestamp * 1000).toLocaleString()}
                          </div>
                          <div className="text-sm text-gray-900 dark:text-white mt-1">
                            {error.message || 'Connection timeout or host unreachable'}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Device Details */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Device Type
                  </label>
                  <div className="text-gray-900 dark:text-white">{device.device_type || 'Unknown'}</div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Property
                  </label>
                  <div className="text-gray-900 dark:text-white">ID: {device.property_id}</div>
                </div>
                {device.tags && device.tags.length > 0 && (
                  <div className="col-span-2">
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                      Tags
                    </label>
                    <div className="flex gap-2 flex-wrap">
                      {device.tags.map((tag: string, i: number) => (
                        <span key={i} className="px-2 py-1 bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded text-sm">
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
