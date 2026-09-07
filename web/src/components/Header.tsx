import { Link as RouterLink, useNavigate } from 'react-router-dom';
import {
  Avatar,
  Badge,
  Box,
  Button,
  Divider,
  Flex,
  Heading,
  HStack,
  Icon,
  Menu,
  MenuButton,
  MenuItem,
  MenuList,
  Spacer,
  Text,
  useColorModeValue,
  useMediaQuery,
} from '@chakra-ui/react';
import { FiBell, FiSearch, FiPlus, FiLogOut, FiUser, FiShield, FiBookmark, FiThumbsUp, FiAtSign, FiMessageSquare, FiFlag } from 'react-icons/fi';
import { useStore } from '../lib/store';
import { useNotifications } from '../lib/useNotifications';

const NOTIF_META: Record<string, { label: string; icon: any; color: string }> = {
  reply: { label: '回复了你的话题', icon: FiMessageSquare, color: '#0d8a6f' },
  mention: { label: '提到了你', icon: FiAtSign, color: '#d97706' },
  like: { label: '赞了你的回复', icon: FiThumbsUp, color: '#dc2626' },
  follow: { label: '关注了你的话题', icon: FiUser, color: '#2563eb' },
  report: { label: '提交了新举报', icon: FiFlag, color: '#9333ea' },
};

export default function Header() {
  const { user, logout } = useStore();
  const { unread, notifs, markRead, markAll } = useNotifications(!!user);
  const nav = useNavigate();
  const [isMobile] = useMediaQuery('(max-width: 768px)');

  // OpenAI Community 风格配色
  const headerBg = useColorModeValue('rgba(244, 245, 247, 0.9)', 'rgba(26, 29, 35, 0.85)');
  const searchBg = useColorModeValue('#ffffff', '#21252c');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const navBg = useColorModeValue('#f7f8fa', '#21252c');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const hoverBg = useColorModeValue('#f3f5f7', '#272c34');
  const menuBg = useColorModeValue('#ffffff', '#1a1d23');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  return (
    <Box
      as="header"
      bg={headerBg}
      backdropFilter="blur(12px) saturate(160%)"
      WebkitBackdropFilter="blur(12px) saturate(160%)"
      borderBottom="1px"
      borderColor={borderColor}
      position="sticky"
      top={0}
      zIndex={100}
    >
      <Flex maxW="1100px" mx="auto" px={4} align="center" h="52px">
        {/* Logo */}
        <Heading
          as={RouterLink}
          to="/"
          size="sm"
          color={textPrimary}
          fontWeight="700"
          letterSpacing="-0.3px"
          fontSize="18px"
        >
          gouache
        </Heading>

        <Spacer />

        {/* 搜索框 */}
        <HStack spacing={2} mr={3}>
          <Box position="relative">
            <Icon
              as={FiSearch}
              position="absolute"
              left={3}
              top="50%"
              transform="translateY(-50%)"
              color={textSecondary}
              boxSize={4}
            />
            <Box
              as="input"
              placeholder="搜索"
              pl={9}
              pr={3}
              py={1.5}
              rounded="md"
              bg={searchBg}
              border="1px"
              borderColor="transparent"
              fontSize="sm"
              color={textPrimary}
              _focus={{ borderColor: accent, outline: 'none' }}
              _placeholder={{ color: textSecondary }}
              w={isMobile ? '100px' : '180px'}
            />
          </Box>
        </HStack>

        {/* 用户区 */}
        {user ? (
          <HStack spacing={2}>
            <Button
              as={RouterLink}
              to="/new"
              leftIcon={<Icon as={FiPlus} boxSize={4} />}
              size="sm"
              bg={accent}
              color="white"
              rounded="md"
              fontWeight="600"
              _hover={{ bg: '#0a7359' }}
            >
              新话题
            </Button>
            {/* 通知铃铛 */}
            <Menu placement="bottom-end">
              <MenuButton as={Button} variant="ghost" rounded="md" p={2} bg="transparent" _hover={{ bg: hoverBg }} pos="relative" minW="auto" h="32px">
                <Icon as={FiBell} boxSize={4} color={textSecondary} />
                {unread > 0 && (
                  <Box pos="absolute" top="-2px" right="-2px" minW="16px" h="16px" rounded="full" bg="#dc2626" color="white" fontSize="10px" fontWeight="700" display="flex" alignItems="center" justifyContent="center" px={0.5} boxShadow="0 1px 3px rgba(0,0,0,.3)">
                    {unread > 99 ? '99+' : unread}
                  </Box>
                )}
              </MenuButton>
              <MenuList bg={menuBg} borderColor={borderColor} shadow="lg" minW="300px" maxH="440px" overflowY="auto" py={0}>
                <Flex align="center" justify="space-between" px={3} py={2.5} borderBottom="1px" borderColor={borderColor}>
                  <Text fontWeight="700" fontSize="sm" color={textPrimary}>通知</Text>
                  {unread > 0 && (
                    <Button size="xs" variant="ghost" color={textSecondary} onClick={markAll} _hover={{ bg: hoverBg }}>全部已读</Button>
                  )}
                </Flex>
                {notifs.length === 0 ? (
                  <Text color={textMuted} fontSize="sm" textAlign="center" py={8}>暂无通知</Text>
                ) : (
                  notifs.slice(0, 20).map((n, idx) => {
                    const meta = NOTIF_META[n.type] || { label: '新动态', icon: FiBell, color: textSecondary };
                    const MIcon = meta.icon;
                    return (
                      <MenuItem
                        key={n.id}
                        onClick={() => {
                          if (!n.read) markRead(n.id);
                          if (n.topic_id) nav(`/topics/${n.topic_id}`);
                        }}
                        bg={n.read ? 'transparent' : hoverBg}
                        color={textPrimary}
                        _hover={{ bg: hoverBg }}
                        py={2.5}
                      >
                        <HStack spacing={3} align="flex-start">
                          <Icon as={MIcon} color={meta.color} boxSize={4} mt={0.5} flexShrink={0} />
                          <Box minW={0}>
                            <Text fontSize="sm" fontWeight="600" noOfLines={2}>{n.content || meta.label}</Text>
                            <Text fontSize="xs" color={textSecondary}>
                              {n.actor?.username || '系统'} {meta.label} · {timeAgo(n.created_at)}
                            </Text>
                          </Box>
                        </HStack>
                      </MenuItem>
                    );
                  })
                )}
              </MenuList>
            </Menu>
            <Menu>
              <MenuButton as={Button} rounded="full" p={0} bg="transparent" _hover={{ bg: hoverBg }} minW="32px" h="32px">
                <Avatar size="xs" name={user.username} bg={accent} />
              </MenuButton>
              <MenuList bg={menuBg} borderColor={borderColor} shadow="md" minW="160px">
                <MenuItem as={RouterLink} to={`/u/${user.username}`} bg={menuBg} color={textPrimary} _hover={{ bg: hoverBg }}>
                  <HStack spacing={2}>
                    <Icon as={FiUser} color={textSecondary} boxSize={4} />
                    <Text fontWeight="600" fontSize="sm">{user.username}</Text>
                  </HStack>
                </MenuItem>
                <MenuItem as={RouterLink} to="/bookmarks" bg={menuBg} color={textPrimary} _hover={{ bg: hoverBg }} fontSize="sm">
                  <HStack spacing={2}>
                    <Icon as={FiBookmark} color={textSecondary} boxSize={4} />
                    <Text>我的收藏</Text>
                  </HStack>
                </MenuItem>
                {user.role === 'admin' && (
                  <MenuItem as={RouterLink} to="/admin" bg={menuBg} color={textPrimary} _hover={{ bg: hoverBg }} fontSize="sm">
                    <HStack spacing={2}>
                      <Icon as={FiShield} color={textSecondary} boxSize={4} />
                      <Text>管理后台</Text>
                    </HStack>
                  </MenuItem>
                )}
                <MenuItem onClick={logout} bg={menuBg} color="#c53030" _hover={{ bg: hoverBg }} fontSize="sm">
                  <HStack spacing={2}>
                    <Icon as={FiLogOut} color="#c53030" boxSize={4} />
                    <Text>退出</Text>
                  </HStack>
                </MenuItem>
              </MenuList>
            </Menu>
          </HStack>
        ) : (
          <HStack spacing={1}>
            <Button as={RouterLink} to="/login" variant="ghost" size="sm" color={textSecondary} _hover={{ bg: hoverBg, color: textPrimary }} fontWeight="500">
              登录
            </Button>
            <Button as={RouterLink} to="/register" size="sm" bg={accent} color="white" rounded="md" fontWeight="600" _hover={{ bg: '#0a7359' }}>
              注册
            </Button>
          </HStack>
        )}
      </Flex>
    </Box>
  );
}

function timeAgo(dateStr?: string) {
  if (!dateStr) return '';
  const now = Date.now();
  const date = new Date(dateStr).getTime();
  const ms = now - date;
  if (ms < 60000) return '刚刚';
  if (ms < 3600000) return `${Math.floor(ms / 60000)}分钟前`;
  if (ms < 86400000) return `${Math.floor(ms / 3600000)}小时前`;
  if (ms < 604800000) return `${Math.floor(ms / 86400000)}天前`;
  return new Date(dateStr).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}