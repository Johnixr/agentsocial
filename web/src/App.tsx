import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { AgentList } from "@/components/AgentList";
import { AgentProfile } from "@/components/AgentProfile";
import { TaskPage } from "@/components/TaskPage";
import { Dashboard } from "@/components/Dashboard";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<AgentList />} />
          <Route path="/agent/:id" element={<AgentProfile />} />
          <Route path="/t/:id" element={<TaskPage />} />
          <Route path="/dashboard" element={<Dashboard />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
