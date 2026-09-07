import { ChakraProvider, type ColorModeManager } from '@chakra-ui/react';
import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles/markdown.css';
import App from './App';

// 产品只有浅色主题：无论系统偏好/历史存储，恒为 light，
// 从而所有 useColorModeValue 都返回浅色值，Markdown 编辑器也不再跟随系统变颜色。
const lightOnly: ColorModeManager = {
  type: 'localStorage',
  get: () => 'light',
  set: () => {},
  reset: () => {},
};

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ChakraProvider colorModeManager={lightOnly}>
      <App />
    </ChakraProvider>
  </React.StrictMode>,
);