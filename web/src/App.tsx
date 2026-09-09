import { useEffect } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { Box } from '@chakra-ui/react';
import Header from './components/Header';
import Home from './pages/Home';
import TopicDetail from './pages/TopicDetail';
import NewTopic from './pages/NewTopic';
import Login from './pages/Login';
import Register from './pages/Register';
import UserProfile from './pages/UserProfile';
import Bookmarks from './pages/Bookmarks';
import AdminPanel from './pages/AdminPanel';
import { useStore } from './lib/store';

// 离开帖子详情页时清空 Header 内的话题标题（Discourse 同款：标题只出现在话题页）。
// 放在路由层保证任何离开路径（点 Logo、后退、点菜单）都复位，而不只依赖组件的卸载清理。
function RouteTitleGuard() {
  const { pathname } = useLocation();
  const setHeaderTitle = useStore((s) => s.setHeaderTitle);
  useEffect(() => {
    if (!/^\/topics\/\d+/.test(pathname)) setHeaderTitle(null);
  }, [pathname, setHeaderTitle]);
  return null;
}

export default function App() {
  return (
    <BrowserRouter>
      <RouteTitleGuard />
      {/* 背景为水粉画水彩晕染（见 markdown.css body），此处透明透出 */}
      <Box minH="100vh">
        <Header />
        <Box maxW="1180px" mx="auto" minH="calc(100vh - 56px)">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/topics/:id" element={<TopicDetail />} />
            <Route path="/new" element={<NewTopic />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/u/:username" element={<UserProfile />} />
            <Route path="/bookmarks" element={<Bookmarks />} />
            <Route path="/admin" element={<AdminPanel />} />
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </Box>
      </Box>
    </BrowserRouter>
  );
}