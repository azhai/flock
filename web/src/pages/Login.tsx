import { useState } from 'react';
import { Link as RouterLink, useLocation, useNavigate } from 'react-router-dom';
import {
  Box,
  Button,
  Flex,
  Heading,
  Icon,
  Input,
  InputGroup,
  InputLeftElement,
  Stack,
  Text,
  useColorModeValue,
  useToast,
} from '@chakra-ui/react';
import { FiLock, FiUser, FiLogIn } from 'react-icons/fi';
import { api } from '../lib/api';
import { useStore } from '../lib/store';

export default function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const setUser = useStore((s) => s.setUser);
  const nav = useNavigate();
  const location = useLocation();
  const toast = useToast();
  const from = (location.state as any)?.from || '/';

  const bgMain = useColorModeValue('#ffffff', '#1a1d23');
  const bgCard = useColorModeValue('#ffffff', '#1a1d23');
  const bgInput = useColorModeValue('#f7f8fa', '#21252c');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  const submit = () => {
    if (!username.trim() || !password.trim()) {
      toast({ title: '请填写用户名和密码', status: 'warning' });
      return;
    }
    api.login({ username, password })
      .then((d: any) => {
        setUser(d.user, d.token);
        toast({ title: '登录成功', status: 'success' });
        nav(from);
      })
      .catch(() => toast({ title: '用户名或密码错误', status: 'error' }));
  };

  return (
    <Flex minH="calc(100vh - 56px)" align="center" justify="center" px={4} bg={bgMain}>
      <Box maxW="md" w="full">
        <Box
          bg={bgCard}
          backdropFilter="blur(16px) saturate(160%)"
          WebkitBackdropFilter="blur(16px) saturate(160%)"
          rounded="lg"
          p={8}
          shadow="xl"
        >
          <Flex align="center" gap={3} mb={8}>
            <Box
              w="10"
              h="10"
              rounded="md"
              bgGradient="linear(to-br, #5B84C4, #7B6FB8)"
              display="flex"
              align="center"
              justify="center"
              shadow="sm"
            >
              <Text color="white" fontWeight="bold" fontSize="xl">G</Text>
            </Box>
            <Heading size="lg" color={textPrimary} fontWeight="700" letterSpacing="-0.5px">
              flock
            </Heading>
          </Flex>

          <Heading size="lg" mb={1} color={textPrimary}>欢迎回来</Heading>
          <Text color={textSecondary} mb={8}>登录你的账号，开始交流分享</Text>

          <Stack spacing={4}>
            <InputGroup>
              <InputLeftElement pointerEvents="none">
                <Icon as={FiUser} color={textSecondary} />
              </InputLeftElement>
              <Input
                placeholder="用户名或邮箱"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                size="lg"
                rounded="md"
                border="1px"
                borderColor={borderColor}
                bg={bgInput}
                color={textPrimary}
                _focus={{ borderColor: accent, boxShadow: 'none' }}
                _placeholder={{ color: textSecondary }}
              />
            </InputGroup>

            <InputGroup>
              <InputLeftElement pointerEvents="none">
                <Icon as={FiLock} color={textSecondary} />
              </InputLeftElement>
              <Input
                type="password"
                placeholder="密码"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && submit()}
                size="lg"
                rounded="md"
                border="1px"
                borderColor={borderColor}
                bg={bgInput}
                color={textPrimary}
                _focus={{ borderColor: accent, boxShadow: 'none' }}
                _placeholder={{ color: textSecondary }}
              />
            </InputGroup>

            <Button
              bg={accent} color="white" _hover={{ bg: '#0a7359' }}
              size="lg"
              onClick={submit}
              rounded="md"
              rightIcon={<FiLogIn />}
            >
              登录
            </Button>
          </Stack>

          <Text mt={6} textAlign="center" color={textSecondary} fontSize="sm">
            还没有账号？
            <RouterLink to="/register">
              <Text as="span" color={accent} fontWeight="600" ml={1}>立即注册</Text>
            </RouterLink>
          </Text>
        </Box>
      </Box>
    </Flex>
  );
}
