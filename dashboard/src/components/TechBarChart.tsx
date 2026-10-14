import { Bar } from 'react-chartjs-2'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js/auto'

interface TechBarChartProps {
  data: Array<{ name: string; count: number }>
}

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip, Legend)

export default function TechBarChart({ data }: TechBarChartProps) {
  // Create modern gradient colors for bars
  const createGradient = (ctx: any, chartArea: any) => {
    const gradient = ctx.createLinearGradient(0, chartArea.bottom, 0, chartArea.top)
    gradient.addColorStop(0, 'rgba(59, 130, 246, 0.6)')
    gradient.addColorStop(0.5, 'rgba(139, 92, 246, 0.7)')
    gradient.addColorStop(1, 'rgba(236, 72, 153, 0.8)')
    return gradient
  }

  const chartData = {
    labels: data.map(t => t.name),
    datasets: [{
      label: 'Projects',
      data: data.map(t => t.count),
      backgroundColor: function(context: any) {
        const chart = context.chart
        const {ctx, chartArea} = chart
        if (!chartArea) return 'rgba(59, 130, 246, 0.6)'
        return createGradient(ctx, chartArea)
      },
      borderColor: 'rgba(139, 92, 246, 1)',
      borderWidth: 2,
      borderRadius: 8,
      borderSkipped: false,
    }],
  }

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        display: false,
      },
      title: {
        display: false,
      },
      tooltip: {
        backgroundColor: 'rgba(15, 23, 42, 0.9)',
        titleColor: 'rgba(255, 255, 255, 1)',
        bodyColor: 'rgba(255, 255, 255, 0.9)',
        borderColor: 'rgba(139, 92, 246, 0.5)',
        borderWidth: 1,
        padding: 12,
        displayColors: false,
        callbacks: {
          label: function(context: any) {
            return ` ${context.parsed.y} projects`
          }
        }
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          color: 'rgba(148, 163, 184, 0.8)',
          font: {
            size: 12,
            weight: '500'
          }
        },
        grid: {
          color: 'rgba(148, 163, 184, 0.1)',
          drawBorder: false,
        },
      },
      x: {
        ticks: {
          color: 'rgba(148, 163, 184, 0.8)',
          font: {
            size: 12,
            weight: '500'
          }
        },
        grid: {
          display: false,
        },
      },
    },
    animation: {
      duration: 1000,
      easing: 'easeOutQuart' as const,
    },
  }

  return (
    <div className="w-full h-80">
      <Bar data={chartData} options={options} />
    </div>
  )
}