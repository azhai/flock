import { useEffect, useState } from 'react';
import { Link as RouterLink, useSearchParams } from 'react-router-dom';
import {
  Avatar,
  Box,
  Button,
  Flex,
  HStack,
  Heading,
  Icon,
  Skeleton,
  Stack,
  Text,
  useColorModeValue,
  useToast,
  VStack,
  Grid,
  GridItem,
} from '@chakra-ui/react';
import { FiMessageSquare, FiMapPin, FiPlus, FiEye, FiArrowUp } from 'react-icons/fi';
import { api } from '../lib/api';
import Sidebar from '../components/Sidebar';
import Pagination from '../components/Pagination';

const SORTS = [
  { key: 'latest', label: '最新' },
  { key: 'hot', label: '活跃' },
  { key: 'top', label: '热门' },
  { key: 'likes', label: '获赞' },
];

export default function Home() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [cats, setCats] = useState<any[]>([]);
  const [topics, setTopics] = useState<any[]>([]);
  const [sort, setSort] = useState(searchParams.get('sort') || 'latest');
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const perPage = 20;
  const toast = useToast();

  const catSlug = searchParams.get('cat');
  const tag = searchParams.get('tag');
  const q = (searchParams.get('q') || '').trim();
  const catId = cats.find((c: any) => c.slug === catSlug)?.id ?? null;

  const bg = useColorModeValue('#ffffff', '#1a1d23');
  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  useEffect(() => {
    api.listCategories().then(setCats).catch(() => toast({ title: '加载分类失败', status: 'error' }));
  }, []);

  useEffect(() => {
    setLoading(true);
    api.listTopics({ category_id: catId, tag, q: q || undefined, sort, page })
      .then((d: any) => {
        setTopics(d.data || []);
        setTotal(d.total || 0);
      })
      .catch(() => toast({ title: '加载话题失败', status: 'error' }))
      .finally(() => setLoading(false));
  }, [catId, tag, q, sort, page]);

  // 搜索词 / 分类 / 标签变化时回到第一页
  useEffect(() => {
    setPage(1);
  }, [q, catSlug, tag]);

  const handleSortChange = (newSort: string) => {
    setSort(newSort);
    const sp = new URLSearchParams(searchParams);
    if (newSort !== 'latest') sp.set('sort', newSort);
    else sp.delete('sort');
    setSearchParams(sp);
    setPage(1);
  };

  return (
    <Grid templateColumns={{ base: '1fr', md: '180px minmax(0, 1fr)' }} gap={0} minH="calc(100vh - 52px)">
      {/* 左栏：分类 + 热门标签 */}
      <GridItem
        bg="transparent"
        borderRight="1px"
        borderColor={borderColor}
        display={{ base: 'none', md: 'block' }}
      >
        <Sidebar />
      </GridItem>

      {/* 主区 */}
      <GridItem bg="transparent">
        <Box px={{ base: 4, md: 6 }} py={4}>
          <Heading size="md" color={textPrimary} fontWeight="700" mb={1}>
            {q
              ? `搜索“${q}”`
              : (cats.find((c: any) => c.slug === catSlug)?.name) || (tag ? `#${tag}` : '最新话题')}
          </Heading>
          <Text fontSize="sm" color={textMuted} mb={4}>
            {total} 个话题 · {cats.length} 个分类
          </Text>

          {/* 下划线 tab */}
          <HStack spacing={1} borderBottom="1px" borderColor={borderColor} mb={0}>
            {SORTS.map((s) => {
              const active = sort === s.key;
              return (
                <Button
                  key={s.key}
                  variant="unstyled"
                  onClick={() => handleSortChange(s.key)}
                  px={3}
                  py={2.5}
                  fontSize="sm"
                  fontWeight={active ? '600' : '500'}
                  color={active ? textPrimary : textSecondary}
                  borderBottom="2px"
                  borderColor={active ? accent : 'transparent'}
                  borderRadius="0"
                  _hover={{ color: textPrimary, bg: bgHover }}
                  mb="-1px"
                >
                  {s.label}
                </Button>
              );
            })}
            <Box flex={1} />
            <Button as={RouterLink} to="/new" size="sm" variant="ghost" leftIcon={<Icon as={FiPlus} boxSize={4} />} color={textSecondary} _hover={{ bg: bgHover, color: textPrimary }} borderRadius="md" mb="6px">
              新话题
            </Button>
          </HStack>

          {/* 话题表格 */}
          {loading ? (
            <Stack spacing={0}>
              {[1, 2, 3, 4, 5].map((i) => (
                <Flex key={i} gap={3} py={3} px={2} align="center">
                  <Box w="22px" flexShrink={0} />
                  <Box flex={1}>
                    <Skeleton h="14px" w="60%" mb={2} />
                    <Skeleton h="11px" w="35%" />
                  </Box>
                </Flex>
              ))}
            </Stack>
          ) : topics.length === 0 ? (
            <Flex direction="column" align="center" justify="center" py={16}>
              <Icon as={FiMessageSquare} boxSize={10} color={textMuted} mb={3} />
              <Text color={textSecondary} fontWeight="600">暂无话题</Text>
              <Text color={textMuted} fontSize="sm" mt={1}>成为第一个发帖的人</Text>
              <Button as={RouterLink} to="/new" size="sm" bg={accent} color="white" mt={4} rounded="md" _hover={{ bg: '#0a7359' }}>
                发布话题
              </Button>
            </Flex>
          ) : (
            <Stack spacing={0} divider={<Box h="1px" bg={borderColor} />}>
              {topics.map((t: any) => {
                const parts = t.participants?.length ? t.participants : (t.user ? [t.user] : []);
                return (
                  <Box key={t.id} as={RouterLink} to={`/topics/${t.id}?back=${encodeURIComponent(`cat=${catSlug || ''}${tag ? `&tag=${tag}` : ''}${sort !== 'latest' ? `&sort=${sort}` : ''}${page > 1 ? `&page=${page}` : ''}`)}`} display="block" px={3} py={4} _hover={{ bg: bgHover }} transition="background 0.12s ease">
                    <Grid templateColumns="minmax(0, 1fr) 52px 52px 110px" gap={4} alignItems="center">
                      <Box minW={0}>
                        <HStack spacing={1.5} mb={1}>
                          {t.pinned && <Icon as={FiMapPin} boxSize={4} color="#d97706" flexShrink={0} />}
                          {t.category && (
                            <Box as="span" fontSize="11px" fontWeight="600" color={t.category.color} bg={t.category.color + '1a'} px={1.5} py={0.5} borderRadius="4px" flexShrink={0}>
                              {t.category.name}
                            </Box>
                          )}
                          <Text fontSize="md" fontWeight="600" color={textPrimary} isTruncated lineHeight="1.4">
                            {t.title}
                          </Text>
                        </HStack>
                        {/* 参与者小头像 + 标签 */}
                        <Flex align="center" gap={3} mt={1.5} flexWrap="wrap">
                          <HStack spacing={1}>
                            {parts.slice(0, 5).map((p: any) => (
                              <Avatar key={p.id} size="xs" name={p.username} bg={accent} w="22px" h="22px" flexShrink={0} />
                            ))}
                          </HStack>
                          {t.tags?.slice(0, 3).map((tagg: any) => (
                            <Box key={tagg.id} as="span" fontSize="11px" fontWeight="600" color={accent} bg={accent + '14'} px={1.5} py={0.5} borderRadius="4px">
                              #{tagg.name}
                            </Box>
                          ))}
                          <Text fontSize="xs" color={textMuted}>
                            {parts[0]?.username || t.user?.username}
                            {t.last_posted_at && ` · ${formatTimeAgo(t.last_posted_at)}`}
                          </Text>
                        </Flex>
                      </Box>
                      <HStack justifyContent="center" spacing={1} display={{ base: 'none', sm: 'flex' }}>
                        <Icon as={FiArrowUp} boxSize={3.5} color={t.like_count > 0 ? accent : textMuted} />
                        <Text fontSize="sm" fontWeight={t.like_count > 0 ? '700' : '500'} color={t.like_count > 0 ? accent : textSecondary}>
                          {t.like_count}
                        </Text>
                      </HStack>
                      <HStack justifyContent="center" spacing={1} display={{ base: 'none', sm: 'flex' }}>
                        <Icon as={FiEye} boxSize={3.5} color={textMuted} />
                        <Text fontSize="sm" color={textSecondary}>
                          {t.views_count}
                        </Text>
                      </HStack>
                      <Text fontSize="sm" color={textMuted} textAlign="right" display={{ base: 'none', md: 'block' }} isTruncated>
                        {t.last_posted_at ? formatTimeAgo(t.last_posted_at) : '—'}
                      </Text>
                    </Grid>
                  </Box>
                );
              })}
            </Stack>
          )}

          {total > perPage && (
            <Pagination page={page} perPage={perPage} total={total} onChange={(p) => { setPage(p); window.scrollTo({ top: 0, behavior: 'smooth' }); }} />
          )}
        </Box>
      </GridItem>
    </Grid>
  );
}

function formatTimeAgo(dateStr: string) {
  const now = new Date();
  const date = new Date(dateStr);
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);
  if (diffMins < 1) return '刚刚';
  if (diffMins < 60) return `${diffMins}分钟前`;
  if (diffHours < 24) return `${diffHours}小时前`;
  if (diffDays < 7) return `${diffDays}天前`;
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
}