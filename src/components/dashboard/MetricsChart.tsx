import React from 'react';
import {
  LineChart,
  Line,
  BarChart,
  Bar,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell
} from 'recharts';

interface ChartDataPoint {
  name: string;
  value: number;
  [key: string]: string | number;
}

interface MetricsChartProps {
  title: string;
  data: ChartDataPoint[];
  type: 'line' | 'bar' | 'area' | 'pie';
  height?: number;
  colors?: string[];
  showGrid?: boolean;
  showLegend?: boolean;
  xAxisLabel?: string;
  yAxisLabel?: string;
  dataKey?: string;
  secondaryDataKey?: string;
}

const MetricsChart: React.FC<MetricsChartProps> = ({
  title,
  data,
  type,
  height = 300,
  colors = ['#3b82f6', '#ef4444', '#10b981', '#f59e0b', '#8b5cf6'],
  showGrid = true,
  showLegend = true,
  xAxisLabel,
  yAxisLabel,
  dataKey = 'value',
  secondaryDataKey
}) => {
  // Generate sample data if no data provided (for development)
  const generateSampleData = (): ChartDataPoint[] => {
    const now = new Date();
    const sampleData: ChartDataPoint[] = [];
    
    for (let i = 0; i < 10; i++) {
      const time = new Date(now.getTime() - (9 - i) * 60000); // Last 10 minutes
      sampleData.push({
        name: time.toLocaleTimeString(),
        value: Math.floor(Math.random() * 100) + 10
      });
    }
    
    return sampleData;
  };

  const chartData = data.length > 0 ? data : generateSampleData();

  const renderChart = () => {
    switch (type) {
      case 'line':
        return (
          <LineChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" />}
            <XAxis 
              dataKey="name" 
              label={xAxisLabel ? { value: xAxisLabel, position: 'bottom' } : undefined}
            />
            <YAxis 
              label={yAxisLabel ? { value: yAxisLabel, angle: -90, position: 'left' } : undefined}
            />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#1f2937', 
                border: '1px solid #374151',
                borderRadius: '8px',
                color: '#f9fafb'
              }}
            />
            {showLegend && <Legend />}
            <Line 
              type="monotone" 
              dataKey={dataKey} 
              stroke={colors[0]} 
              strokeWidth={2}
              dot={{ fill: colors[0], strokeWidth: 2, r: 4 }}
              activeDot={{ r: 6, stroke: colors[0], strokeWidth: 2 }}
            />
            {secondaryDataKey && (
              <Line 
                type="monotone" 
                dataKey={secondaryDataKey} 
                stroke={colors[1]} 
                strokeWidth={2}
                dot={{ fill: colors[1], strokeWidth: 2, r: 4 }}
                activeDot={{ r: 6, stroke: colors[1], strokeWidth: 2 }}
              />
            )}
          </LineChart>
        );

      case 'bar':
        return (
          <BarChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" />}
            <XAxis 
              dataKey="name" 
              label={xAxisLabel ? { value: xAxisLabel, position: 'bottom' } : undefined}
            />
            <YAxis 
              label={yAxisLabel ? { value: yAxisLabel, angle: -90, position: 'left' } : undefined}
            />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#1f2937', 
                border: '1px solid #374151',
                borderRadius: '8px',
                color: '#f9fafb'
              }}
            />
            {showLegend && <Legend />}
            <Bar 
              dataKey={dataKey} 
              fill={colors[0]}
              radius={[4, 4, 0, 0]}
            />
            {secondaryDataKey && (
              <Bar 
                dataKey={secondaryDataKey} 
                fill={colors[1]}
                radius={[4, 4, 0, 0]}
              />
            )}
          </BarChart>
        );

      case 'area':
        return (
          <AreaChart data={chartData}>
            {showGrid && <CartesianGrid strokeDasharray="3 3" />}
            <XAxis 
              dataKey="name" 
              label={xAxisLabel ? { value: xAxisLabel, position: 'bottom' } : undefined}
            />
            <YAxis 
              label={yAxisLabel ? { value: yAxisLabel, angle: -90, position: 'left' } : undefined}
            />
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#1f2937', 
                border: '1px solid #374151',
                borderRadius: '8px',
                color: '#f9fafb'
              }}
            />
            {showLegend && <Legend />}
            <Area 
              type="monotone" 
              dataKey={dataKey} 
              stroke={colors[0]} 
              fill={colors[0]}
              fillOpacity={0.3}
              strokeWidth={2}
            />
            {secondaryDataKey && (
              <Area 
                type="monotone" 
                dataKey={secondaryDataKey} 
                stroke={colors[1]} 
                fill={colors[1]}
                fillOpacity={0.3}
                strokeWidth={2}
              />
            )}
          </AreaChart>
        );

      case 'pie':
        return (
          <PieChart>
            <Pie
              data={chartData}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={(props: { name: string; percent?: number }) => 
                `${props.name} ${props.percent ? (props.percent * 100).toFixed(0) : 0}%`
              }
              outerRadius={80}
              fill="#8884d8"
              dataKey={dataKey}
            >
              {chartData.map((_, index) => (
                <Cell key={`cell-${index}`} fill={colors[index % colors.length]} />
              ))}
            </Pie>
            <Tooltip 
              contentStyle={{ 
                backgroundColor: '#1f2937', 
                border: '1px solid #374151',
                borderRadius: '8px',
                color: '#f9fafb'
              }}
            />
            {showLegend && <Legend />}
          </PieChart>
        );

      default:
        return <div>Unsupported chart type: {type}</div>;
    }
  };



  return (
    <div className="w-full">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
      </div>
      <ResponsiveContainer width="100%" height={height}>
        {renderChart()}
      </ResponsiveContainer>
      {data.length === 0 && (
        <div className="text-center text-sm text-gray-500 mt-2">
          Sample data shown. Connect to live metrics for real-time data.
        </div>
      )}
    </div>
  );
};

export default MetricsChart;
