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
  const chartData = {
    labels: data.map(t => t.name),
    datasets: [{
      label: 'Projects',
      data: data.map(t => t.count),
      backgroundColor: 'rgba(59, 130, 246, 0.6)',
      borderColor: 'rgba(59, 130, 246, 1)',
      borderWidth: 1,
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
        backgroundColor: 'rgba(0, 0, 0, 0.8)',
        titleColor: 'white',
        bodyColor: 'white',
      },
    },
    scales: {
      y: {
        beginAtZero: true,
        ticks: {
          color: 'rgba(156, 163, 175, 1)',
        },
        grid: {
          color: 'rgba(156, 163, 175, 0.2)',
        },
      },
      x: {
        ticks: {
          color: 'rgba(156, 163, 175, 1)',
        },
        grid: {
          display: false,
        },
      },
    },
  }

  return (
    <div style={{ height: '300px' }}>
      <Bar data={chartData} options={options} />
    </div>
  )
}