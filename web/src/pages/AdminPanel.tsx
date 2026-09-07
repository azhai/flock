import { useEffect, useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import {
  Box,
  Button,
  Flex,
  Heading,
  HStack,
  Icon,
  Input,
  Select,
  Spinner,
  Stack,
  Tabs,
  Tab,
  TabList,
  TabPanels,
  TabPanel,
  Text,
  useColorModeValue,
  useToast,
} from '@chakra-ui/react';
import { FiPlus, FiTrash2, FiFlag, FiLock, FiStar, FiUnlock, FiMessageSquare } from 'react-icons/fi';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import { api } from '../lib/api';

const REPORT_STATUS_META: Record<string, { label: string; color: string }> = {
  pending: { label: '待处理', color: '#d97706' },
  resolved: { label: '已处理', color: '#2f855a' },
  dismissed: { label: '已驳回', color: '#8a9099' },
};

export default function AdminPanel() {
  const toast = useToast();
  const [tabIdx, setTabIdx] = useState(0);

  const bg = useColorModeValue('#ffffff', '#1a1d23');
  const bgSub = useColorModeValue('#f7f8fa', '#21252c');
  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  const [stats, setStats] = useState<any>(null);
  const loadStats = () => {
    Promise.all([
      api.listUsers(),
      api.listCategories(),
      api.listTopics({}),
      api.listReports('pending'),
    ])
      .then(([us, cs, ts, rs]) => {
        setStats({
          users: us?.total || 0,
          cats: cs?.length || 0,
          topics: ts?.total || 0,
          reports: rs?.total || 0,
        });
      })
      .catch(() => {});
  };

  useEffect(() => {
    if (tabIdx === 0) loadStats();
    if (tabIdx === 1) loadCategories();
    if (tabIdx === 2) loadReports();
    if (tabIdx === 3) loadTopics();
    if (tabIdx === 4) loadUsers();
  }, [tabIdx]);

  // —— 分类 ——
  const [cats, setCats] = useState<any[]>([]);
  const [newCat, setNewCat] = useState({ name: '', slug: '', description: '' });
  const loadCategories = () => api.listCategories().then(setCats).catch(() => {});
  const createCat = () => {
    if (!newCat.name.trim()) return;
    api.createCategory(newCat).then((c) => {
      setCats([...cats, c]);
      setNewCat({ name: '', slug: '', description: '' });
      toast({ title: '创建成功', status: 'success' });
    }).catch(() => toast({ title: '创建失败', status: 'error' }));
  };
  const delCat = (c: any) => {
    if (!window.confirm(`删除分类「${c.name}」？`)) return;
    api.deleteCategory(c.id).then(() => {
      setCats(cats.filter((x) => x.id !== c.id));
      toast({ title: '已删除', status: 'success' });
    }).catch(() => toast({ title: '删除失败', status: 'error' }));
  };

  // —— 举报流 ——
  const [reports, setReports] = useState<any[]>([]);
  const [reportStatus, setReportStatus] = useState('');
  const loadReports = () =>
    api.listReports(reportStatus || undefined).then((d: any) => setReports(d.data || [])).catch(() => {});
  useEffect(() => {
    if (tabIdx === 2) loadReports();
  }, [reportStatus, tabIdx]);
  const actReport = (r: any, fn: any, msg: string) => {
    fn(r.id).then(() => { loadReports(); toast({ title: msg, status: 'success' }); })
      .catch(() => toast({ title: '操作失败', status: 'error' }));
  };

  // —— 话题 ——
  const [topics, setTopics] = useState<any[]>([]);
  const loadTopics = () => api.listTopics({ sort: 'latest' }).then((d: any) => setTopics(d.data || [])).catch(() => {});
  const topicAct = (t: any, fn: any, msg: string) => {
    fn(t.id).then(() => { loadTopics(); toast({ title: msg, status: 'success' }); })
      .catch(() => toast({ title: '操作失败', status: 'error' }));
  };
  const delTopic = (t: any) => {
    if (!window.confirm(`删除话题「${t.title}」？`)) return;
    api.deleteTopic(t.id).then(() => { loadTopics(); toast({ title: '已删除', status: 'success' }); })
      .catch(() => toast({ title: '删除失败', status: 'error' }));
  };

  // —— 用户 ——
  const [users, setUsers] = useState<any[]>([]);
  const [uq, setUq] = useState('');
  const loadUsers = () => api.listUsers(uq || undefined).then((d: any) => setUsers(d.data || [])).catch(() => {});
  useEffect(() => {
    if (tabIdx === 4) loadUsers();
  }, [uq, tabIdx]);
  const saveUser = (u: any, patch: any, msg: string) => {
    api.adminUpdateUser(u.username, patch).then(() => { loadUsers(); toast({ title: msg, status: 'success' }); })
      .catch(() => toast({ title: '保存失败', status: 'error' }));
  };

  return (
    <Box maxW="1100px" mx="auto" px={{ base: 4, md: 6 }} py={6}>
      <Heading size="lg" color={textPrimary} fontWeight="700" mb={5}>管理后台</Heading>
      <Tabs index={tabIdx} onChange={setTabIdx}>
        <TabList borderColor={borderColor} mb={5}>
          <Tab _selected={{ color: accent, borderColor: accent, fontWeight: '700' }} color={textSecondary}>概览</Tab>
          <Tab _selected={{ color: accent, borderColor: accent, fontWeight: '700' }} color={textSecondary}>分类</Tab>
          <Tab _selected={{ color: accent, borderColor: accent, fontWeight: '700' }} color={textSecondary}>
            <HStack spacing={1}><Text>举报</Text>{stats?.reports > 0 && <Box color="white" bg="#dc2626" rounded="full" px={1.5} minW="16px" textAlign="center" fontSize="10px">{stats.reports}</Box>}</HStack>
          </Tab>
          <Tab _selected={{ color: accent, borderColor: accent, fontWeight: '700' }} color={textSecondary}>话题</Tab>
          <Tab _selected={{ color: accent, borderColor: accent, fontWeight: '700' }} color={textSecondary}>用户</Tab>
        </TabList>

        <TabPanels>
          {/* 概览 */}
          <TabPanel px={0}>
            {!stats ? <Spinner /> : (
              <Stack direction={{ base: 'column', md: 'row' }} spacing={4}>
                {[
                  { label: '用户', value: stats.users, color: '#2563eb', to: 4 },
                  { label: '分类', value: stats.cats, color: '#2f855a', to: 1 },
                  { label: '话题', value: stats.topics, color: '#d97706', to: 3 },
                  { label: '待处理举报', value: stats.reports, color: '#dc2626', to: 2 },
                ].map((s) => (
                  <Box key={s.label} flex={1} bg={bg} border="1px" borderColor={borderColor} rounded="lg" p={5} cursor="pointer" _hover={{ bg: bgHover }} onClick={() => setTabIdx(s.to)} transition="background 0.12s ease">
                    <Text fontSize="3xl" fontWeight="800" color={s.color}>{s.value}</Text>
                    <Text fontSize="sm" color={textMuted}>{s.label}</Text>
                  </Box>
                ))}
              </Stack>
            )}
          </TabPanel>

          {/* 分类 */}
          <TabPanel px={0}>
            <Box bg={bg} border="1px" borderColor={borderColor} rounded="lg" p={5} mb={4}>
              <Heading size="sm" color={textPrimary} mb={4}>新建分类</Heading>
              <Stack direction={{ base: 'column', md: 'row' }} spacing={3}>
                <Input placeholder="名称" value={newCat.name} onChange={(e) => setNewCat({ ...newCat, name: e.target.value })} bg={bgSub} borderColor={borderColor} color={textPrimary} />
                <Input placeholder="slug（可留空）" value={newCat.slug} onChange={(e) => setNewCat({ ...newCat, slug: e.target.value })} bg={bgSub} borderColor={borderColor} color={textPrimary} />
                <Input placeholder="描述" value={newCat.description} onChange={(e) => setNewCat({ ...newCat, description: e.target.value })} bg={bgSub} borderColor={borderColor} color={textPrimary} />
                <Button bg={accent} color="white" _hover={{ bg: '#0a7359' }} leftIcon={<Icon as={FiPlus} boxSize={4} />} onClick={createCat}>创建</Button>
              </Stack>
            </Box>
            <Stack spacing={2}>
              {cats.map((c) => (
                <Flex key={c.id} align="center" gap={3} bg={bg} border="1px" borderColor={borderColor} rounded="md" p={3}>
                  <Box w={2.5} h={2.5} rounded="full" bg={c.color} />
                  <Text fontWeight="600" color={textPrimary}>{c.name}</Text>
                  <Text fontSize="xs" color={textMuted}>/{c.slug}</Text>
                  <Box flex={1} />
                  <Text fontSize="xs" color={textMuted}>{c.topics_count} 话题</Text>
                  <Button size="xs" variant="ghost" color="#c53030" _hover={{ bg: bgHover }} onClick={() => delCat(c)}><Icon as={FiTrash2} boxSize={3} /></Button>
                </Flex>
              ))}
            </Stack>
          </TabPanel>

          {/* 举报流 */}
          <TabPanel px={0}>
            <Flex align="center" gap={3} mb={4}>
              <Select w="160px" value={reportStatus} onChange={(e) => setReportStatus(e.target.value)} bg={bg} borderColor={borderColor} color={textPrimary} size="sm">
                <option value="">全部</option>
                <option value="pending">待处理</option>
                <option value="resolved">已处理</option>
                <option value="dismissed">已驳回</option>
              </Select>
            </Flex>
            <Stack spacing={3}>
              {reports.length === 0 && <Text color={textMuted} fontSize="sm" py={8} textAlign="center">暂无举报</Text>}
              {reports.map((r) => {
                const st = REPORT_STATUS_META[r.status] || REPORT_STATUS_META.pending;
                return (
                  <Box key={r.id} bg={bg} border="1px" borderColor={borderColor} rounded="lg" p={4}>
                    <Flex align="center" gap={2} mb={2} flexWrap="wrap">
                      <Icon as={FiFlag} boxSize={3} color={st.color} />
                      <Text fontSize="sm" fontWeight="700" color={textPrimary}>{r.user?.username} 举报</Text>
                      <Text fontSize="xs" color={textMuted}>#{r.post?.post_number} 楼</Text>
                      <Box fontSize="11px" fontWeight="700" px={1.5} py={0.5} borderRadius="4px" color={st.color} bg={st.color + '14'}>{st.label}</Box>
                      <Box flex={1} />
                      <Text fontSize="xs" color={textMuted}>{new Date(r.created_at || '').toLocaleString('zh-CN')}</Text>
                    </Flex>
                    <Text fontSize="sm" color={textSecondary} bg={bgSub} rounded="md" p={2} mb={3}>理由：{r.reason}</Text>
                    <Box className="markdown-content" color={textSecondary} fontSize="sm" borderLeft="3px" borderColor={borderColor} pl={3} mb={4} maxH="160px" overflow="hidden">
                      <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>{r.post?.raw || ''}</ReactMarkdown>
                    </Box>
                    {r.status === 'pending' && (
                      <HStack spacing={2}>
                        <Button size="xs" colorScheme="green" bg="#2f855a" color="white" _hover={{ bg: '#276749' }} onClick={() => actReport(r, api.resolveReport, '已处理')}>处理（删除违规帖）</Button>
                        <Button size="xs" variant="outline" color={textSecondary} borderColor={borderColor} _hover={{ bg: bgHover }} onClick={() => actReport(r, api.dismissReport, '已驳回')}>驳回</Button>
                      </HStack>
                    )}
                  </Box>
                );
              })}
            </Stack>
          </TabPanel>

          {/* 话题 */}
          <TabPanel px={0}>
            <Stack spacing={2}>
              {topics.map((t) => (
                <Flex key={t.id} align="center" gap={3} bg={bg} border="1px" borderColor={borderColor} rounded="md" p={3}>
                  <Box minW={0} flex={1}>
                    <Text as={RouterLink} to={`/topics/${t.id}`} fontWeight="600" color={textPrimary} isTruncated _hover={{ color: accent }}>{t.title}</Text>
                    <Text fontSize="xs" color={textMuted}>{t.user?.username} · {t.reply_count} 回复 / {t.views_count} 浏览</Text>
                  </Box>
                  <Button size="xs" variant="ghost" color={textSecondary} _hover={{ bg: bgHover }} onClick={() => topicAct(t, t.pinned ? api.unpinTopic : api.pinTopic, t.pinned ? '已取消置顶' : '已置顶')}>
                    <Icon as={t.pinned ? FiUnlock : FiStar} boxSize={3} />
                  </Button>
                  <Button size="xs" variant="ghost" color={textSecondary} _hover={{ bg: bgHover }} onClick={() => topicAct(t, t.closed ? api.openTopic : api.closeTopic, t.closed ? '已解锁' : '已锁定')}>
                    <Icon as={t.closed ? FiMessageSquare : FiLock} boxSize={3} />
                  </Button>
                  <Button size="xs" variant="ghost" color="#c53030" _hover={{ bg: bgHover }} onClick={() => delTopic(t)}><Icon as={FiTrash2} boxSize={3} /></Button>
                </Flex>
              ))}
            </Stack>
          </TabPanel>

          {/* 用户 */}
          <TabPanel px={0}>
            <Input placeholder="搜索用户名 / 邮箱" value={uq} onChange={(e) => setUq(e.target.value)} bg={bg} borderColor={borderColor} color={textPrimary} mb={4} maxW="320px" />
            <Stack spacing={2}>
              {users.map((u) => (
                <Flex key={u.id} align="center" gap={3} bg={bg} border="1px" borderColor={borderColor} rounded="md" p={3} flexWrap="wrap">
                  <Box flex={1} minW="180px">
                    <Text fontWeight="600" color={textPrimary}>{u.username} <Text as={u.role === 'admin' ? 'span' : 'span'} color={textMuted} fontSize="xs">{u.full_name}</Text></Text>
                    <Text fontSize="xs" color={textMuted}>{u.email}</Text>
                  </Box>
                  <Text fontSize="xs" color={textSecondary}>{u.topics_count} 话题 / {u.posts_count} 回复 · Lv{u.level} {u.level_name}</Text>
                  <Select
                    size="xs"
                    w="110px"
                    value={u.role}
                    onChange={(e) => saveUser(u, { role: e.target.value }, '角色已更新')}
                    bg={bgSub} borderColor={borderColor} color={textPrimary}
                  >
                    <option value="user">普通用户</option>
                    <option value="moderator">版主</option>
                    <option value="admin">管理员</option>
                  </Select>
                  <Button
                    size="xs"
                    variant="ghost"
                    color={u.is_active ? '#c53030' : '#2f855a'}
                    _hover={{ bg: bgHover }}
                    onClick={() => saveUser(u, { is_active: !u.is_active }, u.is_active ? '已禁用' : '已启用')}
                  >
                    {u.is_active ? '禁用' : '启用'}
                  </Button>
                </Flex>
              ))}
            </Stack>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </Box>
  );
}