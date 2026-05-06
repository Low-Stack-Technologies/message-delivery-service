import { RouterProvider } from '@tanstack/react-router'
import { AdminStoreProvider } from './admin-store'
import { AdminAuthProvider } from './auth'
import { router } from './router'
import './App.css'

export default function App() {
  return (
    <AdminAuthProvider>
      <AdminStoreProvider>
        <RouterProvider router={router} />
      </AdminStoreProvider>
    </AdminAuthProvider>
  )
}
