import { useEffect, useState } from 'react';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import { Box, Button, Flex, HStack, Heading, Icon, Skeleton, Stack, Text, useColorModeValue, useToast } from '@chakra-ui/react';
import { FiBookmark, FiChevronRight } from 'react-icons/fi';
import { api } from '../lib/api';
import { useStore } from '../lib/store';

export default function Bookmarks() {
  const { user } = useStore();
  const nav = useNavigate();
  const toast = useToast();
  const [bm, setBm] = useState<any[] | null>(null);

  const bg = useColorModeValue('#ffffff', '#1a1d23');
  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  useEffect(() => {
    if (!user) return nav('/login', { state: { from: '/bookmarks' } });
    api.listBookmarks()
      .then((d: any) => setBm(Array.isArray(d) ? d : []))
      .catch(() => toast({ title: '加载收藏失败', status: 'error' }));
  }, [user]);

  return (
    <Box maxW="780px" mx="auto" px={{ base: 4, md: 6 }} py={6}>
      <Heading size="md" color={textPrimary} fontWeight="700" mb={1}>我的收藏</Heading>
      <Text fontSize="sm" color={textMuted} mb={4}>{bm ? `${bm.length} 个话题` : '加载中…'}</Text>

      {!bm ? (
        <Stack spacing={3}>{[1, 2, 3].map((i) => <Skeleton key={i} h="56px" rounded="md" />)}</Stack>
      ) : bm.length === 0 ? (
        <Flex direction="column" align="center" justify="center" py={16} bg={bg} rounded="lg" border="1px" borderColor={borderColor}>
          <Icon as={FiBookmark} boxSize={9} color={textMuted} mb={3} />
          <Text color={textSecondary} fontWeight="600">还没有收藏</Text>
          <Text color={textMuted} fontSize="sm" mt={1}>在话题页点击「收藏」即可加入这里</Text>
          <Button as={RouterLink} to="/" size="sm" bg={accent} color="white" mt={4} rounded="md" _hover={{ bg: '#0a7359' }}>
            去逛逛
          </Button>
        </Flex>
      ) : (
        <Stack spacing={0} divider={<Box h="1px" bg={borderColor} />} bg={bg} rounded="lg" border="1px" borderColor={borderColor} overflow="hidden">
          {bm.map((t: any) => (
            <Box key={t.id} as={RouterLink} to={`/topics/${t.id}`} display="flex" alignItems="center" gap={3} px={4} py={3.5} _hover={{ bg: bgHover }} transition="background 0.12s ease">
              <Box flex={1} minW={0}>
                <HStack spacing={1.5} mb={0.5}>
                  {t.category && (
                    <Box as="span" fontSize="11px" fontWeight="600" color={t.category.color} bg={t.category.color + '1a'} px={1.5} py={0.5} borderRadius="4px" flexShrink={0}>
                      {t.category.name}
                    </Box>
                  )}
                  <Text fontSize="md" fontWeight="600" color={textPrimary} isTruncated>{t.title}</Text>
                </HStack>
                <HStack spacing={1.5}>
                  {t.tags?.slice(0, 3).map((tg: any) => (
                    <Box key={tg.id} as="span" fontSize="11px" fontWeight="500" color={textSecondary} bg={bgHover} px={1.5} py={0.5} borderRadius="4px">#{tg.name}</Box>
                  ))}
                  <Text fontSize="sm" color={textMuted}>{t.posts_count} 回复 · {t.views_count} 浏览</Text>
                </HStack>
              </Box>
              <Icon as={FiChevronRight} boxSize={4} color={textMuted} flexShrink={0} />
            </Box>
          ))}
        </Stack>
      )}
    </Box>
  );
}