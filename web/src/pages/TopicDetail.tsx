import { useEffect, useRef, useState } from 'react';
import { Link as RouterLink, useNavigate, useParams } from 'react-router-dom';
import {
  Avatar,
  Box,
  Button,
  Flex,
  Grid,
  GridItem,
  HStack,
  Heading,
  Icon,
  Input,
  Modal,
  ModalBody,
  ModalCloseButton,
  ModalContent,
  ModalFooter,
  ModalHeader,
  ModalOverlay,
  Select,
  Skeleton,
  Stack,
  Text,
  Textarea,
  useColorModeValue,
  useToast,
} from '@chakra-ui/react';
import { FiMessageSquare, FiLock, FiSend, FiMapPin, FiArrowLeft, FiBookmark, FiEdit2, FiTrash2, FiFlag, FiX, FiArrowUp, FiArrowDown, FiBell } from 'react-icons/fi';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkBreaks from 'remark-breaks';
import MDEditor from '@uiw/react-md-editor/nohighlight';
import '@uiw/react-md-editor/markdown-editor.css';
import CodeBlock from '../components/CodeBlock';
import { TagInput, TagChips } from '../components/tag-input';
import { ensureSyntaxHighlight } from '../lib/highlightProvider';
import { api } from '../lib/api';
import { useStore } from '../lib/store';
import Sidebar from '../components/Sidebar';

export default function TopicDetail() {
  const { id } = useParams();
  const nav = useNavigate();
  const { user } = useStore();
  const [topic, setTopic] = useState<any>(null);
  const [posts, setPosts] = useState<any[]>([]);
  const [reply, setReply] = useState('');
  const [loading, setLoading] = useState(true);
  const [cats, setCats] = useState<any[]>([]);
  const [editOpen, setEditOpen] = useState(false);
  const [eTitle, setETitle] = useState('');
  const [eCat, setECat] = useState('');
  const [eTags, setETags] = useState<string[]>([]);
  const toast = useToast();
  const [isDark] = useState(() => false); // 单一浅色主题，编辑器始终 light

  const bg = useColorModeValue('#ffffff', '#1a1d23');
  const bgSub = useColorModeValue('#f7f8fa', '#21252c');
  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    api.getTopic(id)
      .then(setTopic)
      .catch(() => toast({ title: '话题不存在', status: 'error' }))
      .finally(() => setLoading(false));
    api.listPosts(id)
      .then((d: any) => setPosts(d.data || []))
      .catch(() => {});
  }, [id]);

  // 提前懒加载语法高亮器，保证打开 Preview 时已就绪
  useEffect(() => {
    ensureSyntaxHighlight().catch(() => {});
    api.listCategories().then(setCats).catch(() => {});
  }, []);

  const openEditTopic = () => {
    setETitle(topic?.title || '');
    setECat(topic?.category_id ? String(topic.category_id) : '');
    setETags((topic?.tags || []).map((x: any) => x.name));
    setEditOpen(true);
  };

  const saveEditTopic = () => {
    api.updateTopic(id, {
      title: eTitle.trim(),
      category_id: eCat ? Number(eCat) : null,
      tags: eTags,
    })
      .then((t: any) => {
        setTopic(t);
        setEditOpen(false);
        toast({ title: '话题已更新', status: 'success' });
      })
      .catch(() => toast({ title: '更新失败', status: 'error' }));
  };

  const isStaff = !!user && (user.role === 'admin' || user.role === 'moderator');
  const canEditTopic = !!user && (isStaff || user.username === topic?.user?.username);

  const submitReply = () => {
    if (!user) return nav('/login', { state: { from: `/topics/${id}` } });
    if (!reply.trim()) return;
    api.createPost(id, { raw: reply })
      .then((p: any) => {
        setPosts([...posts, p]);
        setReply('');
        toast({ title: '回复成功', status: 'success' });
      })
      .catch(() => toast({ title: '回复失败', status: 'error' }));
  };

  const updatePostBody = (p: any, raw: string) => {
    api.updatePost(p.id, { raw })
      .then((d: any) => {
        setPosts(posts.map((x) => (x.id === p.id ? { ...x, raw: d.raw, editing: false } : x)));
        toast({ title: '已更新', status: 'success' });
      })
      .catch(() => toast({ title: '保存失败', status: 'error' }));
  };

  const deletePostBody = (p: any) => {
    if (!window.confirm('确定删除这条回复？')) return;
    api.deletePost(p.id)
      .then(() => {
        setPosts(posts.filter((x) => x.id !== p.id));
        toast({ title: '已删除', status: 'success' });
      })
      .catch(() => toast({ title: '删除失败', status: 'error' }));
  };

  const toggleBookmark = () => {
    if (!user) return nav('/login', { state: { from: `/topics/${id}` } });
    api.bookmarkTopic(id)
      .then((d: any) => setTopic((t: any) => (t ? { ...t, bookmarked: d.bookmarked } : t)))
      .catch(() => toast({ title: '操作失败', status: 'error' }));
  };

  const toggleFollow = () => {
    if (!user) return nav('/login', { state: { from: `/topics/${id}` } });
    api.followTopic(id)
      .then((d: any) => setTopic((t: any) => (t ? { ...t, followed: d.followed } : t)))
      .catch(() => toast({ title: '操作失败', status: 'error' }));
  };

  const votePost = (p: any, vote: number) => {
    if (!user) return nav('/login', { state: { from: `/topics/${id}` } });
    const target = p.my_vote === vote ? 0 : vote; // 再点一次取消
    api.votePost(p.id, target)
      .then((d: any) => {
        setPosts(posts.map((x) => (x.id === p.id ? { ...x, score: d.score, my_vote: d.my_vote, like_count: d.like_count } : x)));
      })
      .catch(() => toast({ title: '投票失败', status: 'error' }));
  };

  const [reportTarget, setReportTarget] = useState<any>(null);
  const [reportReason, setReportReason] = useState('');
  const submitReport = () => {
    if (!reportTarget) return;
    if (!reportReason.trim()) return toast({ title: '请填写举报理由', status: 'warning' });
    api.reportPost(reportTarget.id, reportReason)
      .then(() => toast({ title: '举报已提交，感谢反馈', status: 'success' }))
      .catch(() => toast({ title: '举报失败', status: 'error' }));
    setReportTarget(null);
    setReportReason('');
  };

  const firstPost = posts.find((p: any) => p.post_number === 1);
  const allPosts = firstPost ? posts : posts; // 渲染顺序 = 帖序号顺序
  // ── 右侧帖子节点导航：滚动容器 + 每帖 ref ──────────────
  const scrollRef = useRef<any>(null);
  const postRefs = useRef<Record<string, HTMLElement>>({});
  const [nodeTops, setNodeTops] = useState<number[]>([]);
  const [activePostId, setActivePostId] = useState<string | null>(null);

  const registerPost = (id: string) => (el: any) => {
    if (el) postRefs.current[id] = el;
  };

  const syncNodes = () => {
    const el = scrollRef.current;
    if (!el || !allPosts.length) return;
    const total = el.scrollHeight || 1;
    setNodeTops(allPosts.map((p: any) => ((postRefs.current[p.id]?.offsetTop || 0) / total) * 100));
  };

  const handleScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    const st = el.scrollTop + 60;
    let cur: string | null = null;
    for (const p of allPosts) {
      const pe = postRefs.current[p.id];
      if (pe && pe.offsetTop <= st) cur = p.id;
    }
    setActivePostId(cur || allPosts[0]?.id || null);
    syncProgress();
    maybeSnap(); // 滚轮滚动到刻度附近时吸附
  };

  const jumpToPost = (pid: string) => {
    postRefs.current[pid]?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  };

  // ── 右侧钓鱼线刻度式滚动条：进度 + 拖拽 ──────────────
  const railRef = useRef<any>(null);
  const [prog, setProg] = useState(0);          // 可视区在整条中的相对位置 0..1
  const [railH, setRailH] = useState(200);
  const draggingRef = useRef(false);

  const syncProgress = () => {
    const el = scrollRef.current;
    if (!el) return;
    const max = el.scrollHeight - el.clientHeight;
    setProg(max > 0 ? el.scrollTop / max : 0);
  };

  // 把某个指针 y 换算成内容滚动位置（拖拽滑块）
  const setThumbFromY = (clientY: number) => {
    const rail = railRef.current, el = scrollRef.current;
    if (!rail || !el) return;
    const rect = rail.getBoundingClientRect();
    const rh = rect.height;
    const max = el.scrollHeight - el.clientHeight;
    if (max <= 0) return;
    const th = Math.max(rh * (el.clientHeight / el.scrollHeight), 18);
    let t = (clientY - rect.top - th / 2) / (rh - th);
    el.scrollTop = Math.min(Math.max(t, 0), 1) * max; // 程序赋值会触发 onScroll → 自动 sync 进度/高亮
  };

  const onRailPointerDown = (e: any) => {
    draggingRef.current = true;
    const rail = railRef.current;
    (rail as any).setPointerCapture?.(e.pointerId);
    setThumbFromY(e.clientY);
  };
  const onRailPointerMove = (e: any) => {
    if (!draggingRef.current) return;
    setThumbFromY(e.clientY);
  };
  // 鼠标停靠在滑块/滚条上时，滚轮也控制帖子滚动
  const onRailWheel = (e: any) => {
    const el = scrollRef.current;
    if (!el) return;
    e.preventDefault?.();
    el.scrollTop = Math.max(0, el.scrollTop + (e.deltaY || 0));
  };
  const onRailPointerUp = () => {
    draggingRef.current = false;
    maybeSnap(); // 松开时对齐到最近的刻度
  };

  // ── 刻度吸附：接近某帖时把可视区贴靠到该帖，滚离一段距离后才释放，
  //    产生"钓鱼线结/刻度上车钩"式的分段阻力手感 ──────────────────
  const snapRef = useRef<{ pid: string; pos: number } | null>(null);
  const maybeSnap = () => {
    const el = scrollRef.current;
    if (!el || draggingRef.current) return;
    const SNAP = 64;      // 距刻度多少像素内触发吸附
    const RELEASE = 160;  // 滚离当前吸附帖多少像素后才允许吸附其他刻度
    const cur = el.scrollTop;
    let best: { pid: string; pos: number } | null = null;
    let bestD = SNAP;
    for (const p of allPosts) {
      const pe = postRefs.current[p.id];
      if (!pe) continue;
      const target = Math.max(pe.offsetTop - 24, 0); // 帖顶留一点空白
      const d = Math.abs(cur - target);
      if (d < bestD) { bestD = d; best = { pid: p.id, pos: target }; }
    }
    const snap = snapRef.current;
    if (best) {
      // 仍吸附在上一刻度附近：保持，不重复赋值（避免来回抖动）
      if (snap && Math.abs(cur - snap.pos) < RELEASE) return;
      snapRef.current = best;
      el.scrollTop = best.pos;
    } else if (snap) {
      snapRef.current = null; // 远离所有刻度，解除吸附
    }
  };

  useEffect(() => {
    const m = () => { if (railRef.current) setRailH(railRef.current.clientHeight); };
    m();
    window.addEventListener('resize', m);
    return () => window.removeEventListener('resize', m);
  }, []);

  useEffect(() => {
    if (!allPosts.length) return;
    syncNodes();
    handleScroll();
    const t = setTimeout(() => {
      if (railRef.current) setRailH(railRef.current.clientHeight); // 加载后才有真实列高
      syncNodes();
      syncProgress();
    }, 300); // 等排版/图片稳定后校准
    return () => clearTimeout(t);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [posts]);

  // 打开详情/切换 id 时：回到顶部并把浏览焦点落在主帖内容区，滚轮立即可用
  useEffect(() => {
    const el = scrollRef.current;
    if (!el) return;
    el.scrollTop = 0;
    setProg(0);
    el.focus({ preventScroll: true });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const setFlag = (fn: any, okMsg: string) => {
    fn(id)
      .then(() => {
        toast({ title: okMsg, status: 'success' });
        api.getTopic(id).then(setTopic);
      })
      .catch(() => toast({ title: '操作失败', status: 'error' }));
  };

  const confirmDeleteTopic = () => {
    if (!window.confirm('确定删除整个话题（含所有回复）？此操作不可撤销。')) return;
    api.deleteTopic(id)
      .then(() => nav('/'))
      .catch(() => toast({ title: '删除失败', status: 'error' }));
  };

  // 滑块几何：可视区高度占比对应的滑块大小与位置（右侧钓鱼线刻度条）
  const scEl = scrollRef.current;
  const scH = scEl?.scrollHeight || 1;
  const cliH = scEl?.clientHeight || scH;
  const thumbHeight = Math.max(railH * (cliH / scH), 18);
  const thumbTop = prog * Math.max(railH - thumbHeight, 0);

  if (loading) {
    return (
      <Box maxW="780px" mx="auto" px={4} py={6}>
        <Stack spacing={3}>
          {[1, 2, 3].map((i) => <Skeleton key={i} h="80px" rounded="md" />)}
        </Stack>
      </Box>
    );
  }

  if (!topic) {
    return (
      <Flex direction="column" align="center" justify="center" py={20}>
        <Text color={textSecondary}>话题不存在</Text>
        <Button as={RouterLink} to="/" size="sm" variant="outline" mt={4} rounded="md">返回首页</Button>
      </Flex>
    );
  }

  return (
    <Box maxW="1180px" mx="auto">
      <Grid templateColumns={{ base: '1fr', md: '180px minmax(0, 1fr) 110px' }} gap={0} alignItems="start">
      {/* 左侧：分类 + 热门标签，保持与首页一致，跳转后选中状态不变 */}
      <Box
        display={{ base: 'none', md: 'block' }}
        borderRight="1px"
        borderColor={borderColor}
        pos="sticky"
        top="52px"
        maxH="calc(100vh - 52px)"
        overflowY="auto"
      >
        <Sidebar />
      </Box>
      <Box minW={0} px={{ base: 4, md: 6 }} py={6} display="flex" flexDirection="column" maxH="calc(100vh - 52px)">
      {/* 话题头部 */}
      <Box mb={6} flexShrink={0}>
        <HStack spacing={2} mb={3} flexWrap="wrap">
          {topic.pinned && (
            <HStack spacing={1} fontSize="11px" fontWeight="600" color="#d97706" bg="#fef3c7" px={2} py={1} borderRadius="4px">
              <Icon as={FiMapPin} boxSize={3} /> 置顶
            </HStack>
          )}
          {topic.closed && (
            <HStack spacing={1} fontSize="11px" fontWeight="600" color="#c53030" bg="#fed7d7" px={2} py={1} borderRadius="4px">
              <Icon as={FiLock} boxSize={3} /> 已锁定
            </HStack>
          )}
          {topic.category && (
            <Box fontSize="11px" fontWeight="600" color={topic.category.color} bg={topic.category.color + '1a'} px={2} py={1} borderRadius="4px">
              {topic.category.name}
            </Box>
          )}
          {topic.tags?.map((tag: any) => (
            <Box as={RouterLink} key={tag.id} to={`/?tag=${tag.slug}`} fontSize="11px" fontWeight="600" color={textSecondary} bg={bgHover} px={2} py={1} borderRadius="4px" _hover={{ bg: accent + '1a', color: textPrimary }}>
              #{tag.name}
            </Box>
          ))}
        </HStack>
        <Heading size="lg" color={textPrimary} fontWeight="700" lineHeight="1.4" mb={3}>
          {topic.title}
        </Heading>
        <Flex align="center" justify="space-between" flexWrap="wrap" gap={2}>
          <HStack spacing={3} fontSize="sm" color={textMuted}>
            <Avatar size="xs" name={topic.user?.username} bg={accent} />
            <Text fontWeight="600" color={textSecondary}>{topic.user?.username}</Text>
            <Text>· {formatDate(topic.created_at)}</Text>
            <Text>· {topic.views_count} 浏览</Text>
            <Text>· {topic.posts_count} 回复</Text>
          </HStack>
          <HStack spacing={1}>
            <Button size="sm" variant="ghost" leftIcon={<Icon as={FiBookmark} boxSize={4} color={topic.bookmarked ? accent : textSecondary} />} color={topic.bookmarked ? accent : textSecondary} _hover={{ bg: bgHover }} onClick={toggleBookmark}>
              {topic.bookmarked ? '已收藏' : '收藏'}
            </Button>
            <Button size="sm" variant="ghost" leftIcon={<Icon as={FiBell} boxSize={4} color={topic.followed ? accent : textSecondary} />} color={topic.followed ? accent : textSecondary} _hover={{ bg: bgHover }} onClick={toggleFollow}>
              {topic.followed ? '已关注' : '关注'}
            </Button>
            {canEditTopic && (
              <>
                <Button size="sm" variant="ghost" leftIcon={<Icon as={FiEdit2} boxSize={4} />} color={textSecondary} _hover={{ bg: bgHover, color: textPrimary }} onClick={openEditTopic}>
                  编辑
                </Button>
                <Button size="sm" variant="ghost" color={textSecondary} _hover={{ bg: bgHover }} onClick={() => setFlag(topic.pinned ? api.unpinTopic : api.pinTopic, topic.pinned ? '已取消置顶' : '已置顶')}>
                  <Icon as={FiFlag} boxSize={4} />
                </Button>
                <Button size="sm" variant="ghost" color={textSecondary} _hover={{ bg: bgHover }} onClick={() => setFlag(topic.closed ? api.openTopic : api.closeTopic, topic.closed ? '已解锁' : '已锁定')}>
                  <Icon as={topic.closed ? FiMessageSquare : FiLock} boxSize={4} />
                </Button>
                <Button size="sm" variant="ghost" color="#c53030" _hover={{ bg: bgHover }} onClick={confirmDeleteTopic}>
                  <Icon as={FiTrash2} boxSize={4} />
                </Button>
              </>
            )}
          </HStack>
        </Flex>
      </Box>

      {/* 内容区：主帖 + 回复。长帖或多回复时在区域内滚动；显示右侧粗滚动条，滚轮可直接滚动 */}
      <Box
        ref={scrollRef}
        onScroll={handleScroll}
        pos="relative"
        overflowY="auto"
        flex="1"
        minH={0}
        pr={{ base: 1, md: 1 }}
        tabIndex={-1}
        outline="none"
        className="forum-thin-scroll"
      >
      {/* 帖子流：主贴直接显示，回复用卡片分隔。每帖包一层 div 供右侧节点定位 */}
      {firstPost && (
        <div key={firstPost.id} ref={registerPost(firstPost.id)}>
        <PostCard
          post={firstPost}
          isOp
          user={user}
          isStaff={isStaff}
          canEdit={!!user && (isStaff || user.username === firstPost.user?.username)}
          updatePostBody={updatePostBody}
          deletePostBody={deletePostBody}
          onVote={votePost}
          onReport={setReportTarget}
          borderColor={borderColor}
          bgHover={bgHover}
          textPrimary={textPrimary}
          textSecondary={textSecondary}
          textMuted={textMuted}
          accent={accent}
          isDark={isDark}
        />
        </div>
      )}
      <Stack spacing={0} mt={firstPost ? 0 : 1} divider={<Box h="1px" bg={borderColor} />}>
        {posts.filter((p: any) => p.post_number !== 1).map((p: any) => (
          <div key={p.id} ref={registerPost(p.id)}>
          <PostCard
            post={p}
            user={user}
            isStaff={isStaff}
            canEdit={!!user && (isStaff || user.username === p.user?.username)}
            updatePostBody={updatePostBody}
            deletePostBody={deletePostBody}
            onVote={votePost}
            onReport={setReportTarget}
            borderColor={borderColor}
            bgHover={bgHover}
            textPrimary={textPrimary}
            textSecondary={textSecondary}
            textMuted={textMuted}
            accent={accent}
            isDark={isDark}
          />
          </div>
        ))}
      </Stack>
      </Box>

      {/* 回复框：固定在底部，不随帖子和回复一起滚动 */}
      <Box flexShrink={0} mt={6} bg={bgSub} rounded="md" p={4} border="1px" borderColor={borderColor}>
        {topic.closed ? (
          <Text color="#c53030" fontSize="sm" fontWeight="600">话题已锁定，无法回复</Text>
        ) : (
          <>
            <Box data-color-mode={isDark ? 'dark' : 'light'} mb={3}>
              <MDEditor
                value={reply}
                onChange={(v) => setReply(v || '')}
                height={120}
                preview="edit"
                previewOptions={{ components: { code: CodeBlock } }}
                textareaProps={{ placeholder: '写下你的回复…（支持 Markdown，单回车自动换行）' }}
              />
            </Box>
            <Flex justify="flex-end">
              <Button
                size="sm"
                bg={accent}
                color="white"
                onClick={submitReply}
                rightIcon={<Icon as={FiSend} boxSize={3} />}
                rounded="md"
                _hover={{ bg: '#0a7359' }}
                fontWeight="600"
              >
                回复
              </Button>
            </Flex>
          </>
        )}
      </Box>
      </Box>

      {/* 右侧：返回按钮 + 钓鱼线刻度式滚动条。刻度即各帖/回复，滑块显示当前可视区；拖拽/点击滚动 */}
      <Box as="aside" display={{ base: 'none', md: 'flex' }} flexDirection="column" alignItems="center" pos="sticky" top="52px" maxH="calc(100vh - 52px)" pl={3} pr={3} pt={2} pb={2} borderLeft="1px" borderColor={borderColor}>
        <Flex justify="flex-end" w="full" mb={3}>
          <Button as={RouterLink} to="/" variant="ghost" size="sm" color={textSecondary} _hover={{ bg: bgHover, color: textPrimary }} px={1}>
            <Icon as={FiArrowLeft} boxSize={4} /> 返回
          </Button>
        </Flex>
        <Box
          ref={railRef}
          flex="1"
          pos="relative"
          w="26px"
          minH="160px"
          cursor="grab"
          userSelect="none"
          touchAction="none"
          _active={{ cursor: 'grabbing' }}
          onPointerDown={onRailPointerDown}
          onPointerMove={onRailPointerMove}
          onPointerUp={onRailPointerUp}
          onPointerCancel={onRailPointerUp}
          onWheel={onRailWheel}
        >
          {/* 轨道线（细绳），即"钓鱼线"本体 */}
          <Box pos="absolute" top="0" bottom="0" left="50%" w="3px" transform="translateX(-1.5px)" borderRadius="full" bg={borderColor} opacity="0.85" />
          {/* 刻度：小圆点，按各帖/回复的实际落点分布；点按直接跳转 */}
          {allPosts.map((p: any, i: number) => {
            if (i === 0) return null; // 去掉顶部（主贴）那个圆点
            const cur = (activePostId ?? allPosts[0]?.id ?? null) === p.id;
            return (
              <Box
                key={p.id}
                pos="absolute"
                left="50%"
                top={(nodeTops[i] ?? 0) + '%'}
                transform="translate(-50%,-50%)"
                w={cur ? '12px' : '9px'}
                h={cur ? '12px' : '9px'}
                borderRadius="full"
                bg={cur ? accent : textMuted}
                opacity={cur ? 1 : 0.55}
                cursor="pointer"
                title={p.post_number === 1 ? '主贴' : `回复 #${p.post_number}`}
                _hover={{ bg: accent, opacity: 1 }}
                onPointerDown={(e: any) => e.stopPropagation()}
                onClick={() => jumpToPost(p.id)}
                style={{ touchAction: 'none' }}
              />
            );
          })}
          {/* 拖拽滑块：高亮当前可视区，划过刻度时呈现"吃力/分段"的手感 */}
          <Box
            pos="absolute"
            left="50%"
            transform="translateX(-50%)"
            w="6px"
            h={`${thumbHeight}px`}
            top={`${thumbTop}px`}
            borderRadius="full"
            bg={accent}
            opacity="0.9"
            boxShadow={`0 0 6px ${accent + '55'}`}
            transition="top 0.08s ease-out"
            pointerEvents="none"
          />
        </Box>
        {/* 当前所在帖子标签 */}
        <Text fontSize="xs" color={textMuted} textAlign="center" pt={3} isTruncated>
          {(() => {
            const p = allPosts.find((x: any) => x.id === (activePostId ?? allPosts[0]?.id ?? null));
            return p ? (p.post_number === 1 ? '主贴' : '回复 #' + p.post_number) : '';
          })()}
        </Text>
      </Box>
      </Grid>

      {/* 举报对话框 */}
      <Modal isOpen={!!reportTarget} onClose={() => setReportTarget(null)} isCentered>
        <ModalOverlay />
        <ModalContent bg={bg} color={textPrimary}>
          <ModalHeader fontSize="md" fontWeight="700">举报这条回复</ModalHeader>
          <ModalCloseButton />
          <ModalBody>
            <Textarea
              placeholder="请简要说明举报理由（垃圾信息 / 引战 / 不友善 …）"
              value={reportReason}
              onChange={(e) => setReportReason(e.target.value)}
              rows={3}
              bg={bgSub}
              borderColor={borderColor}
            />
          </ModalBody>
          <ModalFooter>
            <Button size="sm" variant="ghost" mr={3} color={textSecondary} onClick={() => setReportTarget(null)}>取消</Button>
            <Button size="sm" colorScheme="red" bg="#c53030" color="white" _hover={{ bg: '#9b2c2c' }} onClick={submitReport} isDisabled={!reportReason.trim()}>
              提交举报
            </Button>
          </ModalFooter>
        </ModalContent>
      </Modal>

      {/* 编辑话题模态框 */}
      <Modal isOpen={editOpen} onClose={() => setEditOpen(false)} size="xl">
        <ModalOverlay />
        <ModalContent bg={bg}>
          <ModalHeader>编辑话题</ModalHeader>
          <ModalCloseButton />
          <ModalBody pb={6}>
            <Stack spacing={4}>
              <Box>
                <Text fontWeight="600" mb={1} fontSize="sm" color={textPrimary}>标题</Text>
                <Input
                  value={eTitle}
                  onChange={(e) => setETitle(e.target.value)}
                  placeholder="标题：一句话说清你的话题"
                  size="md"
                  rounded="md"
                  border="1px"
                  borderColor={borderColor}
                  bg={bgSub}
                  color={textPrimary}
                  _focus={{ borderColor: accent, boxShadow: 'none' }}
                />
              </Box>

              <Grid templateColumns={{ base: '1fr', sm: '2fr 3fr' }} gap={4}>
                <Box>
                  <Text fontWeight="600" mb={1} fontSize="sm" color={textPrimary}>分类</Text>
                  <Select
                    value={eCat}
                    onChange={(e) => setECat(e.target.value)}
                    rounded="md"
                    border="1px"
                    borderColor={borderColor}
                    bg={bgSub}
                    color={textPrimary}
                    size="md"
                    _focus={{ borderColor: accent, boxShadow: 'none' }}
                  >
                    <option value="">选择分类（可选）</option>
                    {cats.map((c: any) => (
                      <option key={c.id} value={c.id}>{c.name}</option>
                    ))}
                  </Select>
                </Box>

                <Box>
                  <Text fontWeight="600" mb={1} fontSize="sm" color={textPrimary}>标签</Text>
                  <TagInput value={eTags} onChange={setETags} accent={accent} borderColor={borderColor} bgInput={bgSub} textPrimary={textPrimary} textSecondary={textSecondary} />
                </Box>
              </Grid>

              <Box>
                <Text fontWeight="600" mb={1} fontSize="sm" color={textPrimary}>已选标签</Text>
                <TagChips value={eTags} onChange={setETags} accent={accent} textSecondary={textSecondary} />
              </Box>
            </Stack>
          </ModalBody>

          <ModalFooter>
            <Button size="sm" variant="ghost" mr={3} color={textSecondary} onClick={() => setEditOpen(false)}>取消</Button>
            <Button size="sm" bg={accent} color="white" _hover={{ bg: '#0a7359' }} onClick={saveEditTopic} isDisabled={!eTitle.trim()}>保存</Button>
          </ModalFooter>
        </ModalContent>
      </Modal>
    </Box>
  );
}

function PostCard({ post, user, isOp, canEdit, updatePostBody, deletePostBody, onVote, onReport, borderColor, bgHover, textPrimary, textSecondary, textMuted, accent, isDark }: any) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(post.raw);
  const lv = LEVEL_META[post.user?.level] || { name: (post.user?.level_name || '新手'), color: textMuted, bg: 'transparent' };

  const actions = (
    <Flex align="center" mt={3} flexWrap="wrap" gap={1.5}>
      {/* 顶/踩：评分居中，一套系统；再点取消当前票 */}
      <HStack border="1px" borderColor={borderColor} borderRadius="md" spacing={0} overflow="hidden" bg={bgHover}>
        <Button size="xs" variant="ghost" h={7} px={2} color={post.my_vote === 1 ? accent : textMuted} _hover={{ bg: accent + '14' }} onClick={() => onVote(post, 1)} aria-label="顶">
          <Icon as={FiArrowUp} boxSize={4} />
        </Button>
        <Text w="42px" textAlign="center" fontSize="sm" fontWeight="700" color={post.like_count > 0 ? accent : post.like_count < 0 ? '#c53030' : textSecondary}>
          {post.like_count ?? 0}
        </Text>
        <Button size="xs" variant="ghost" h={7} px={2} color={post.my_vote === -1 ? '#c53030' : textMuted} _hover={{ bg: '#fed7d714' }} onClick={() => onVote(post, -1)} aria-label="踩">
          <Icon as={FiArrowDown} boxSize={4} />
        </Button>
      </HStack>
      {!canEdit && !isOp && (
        <Button size="xs" variant="ghost" color={textMuted} _hover={{ bg: bgHover, color: '#c53030' }} leftIcon={<Icon as={FiFlag} boxSize={3} />} onClick={() => onReport(post)}>
          举报
        </Button>
      )}
      {canEdit && (
        <Button size="xs" variant="ghost" color={textMuted} _hover={{ bg: bgHover, color: textPrimary }} leftIcon={<Icon as={FiEdit2} boxSize={3} />} onClick={() => { setEditing(true); setDraft(post.raw); }}>
          编辑
        </Button>
      )}
      {canEdit && (
        <Button size="xs" variant="ghost" color="#c53030" _hover={{ bg: bgHover }} leftIcon={<Icon as={FiTrash2} boxSize={3} />} onClick={() => deletePostBody(post)}>
          删除
        </Button>
      )}
    </Flex>
  );

  // 主贴：直接显示，只用小头像标注作者行，不带回复卡片框
  if (isOp) {
    return (
      <Box pb={6}>
      {editing ? (
        <Box pl={3} pr={3}>
          <Box data-color-mode={isDark ? 'dark' : 'light'} mb={2}>
            <MDEditor value={draft} onChange={(v) => setDraft(v || '')} height={220} preview="edit" />
          </Box>
          <Flex gap={2} justify="flex-end">
            <Button size="xs" variant="ghost" color={textMuted} onClick={() => setEditing(false)}>取消</Button>
            <Button size="xs" colorScheme="teal" bg={accent} color="white" _hover={{ bg: '#0a7359' }} onClick={() => { updatePostBody(post, draft); setEditing(false); }}>保存</Button>
          </Flex>
        </Box>
      ) : (
        <>
          <Box className="markdown-content" color={textPrimary} fontSize="md" lineHeight="1.75">
            <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]} components={{ code: CodeBlock }}>{post.raw}</ReactMarkdown>
          </Box>
          {actions}
        </>
      )}
    </Box>
  );
}

  return (
    <Box py={5} px={2}>
      <HStack spacing={3} mb={1} align="flex-start">
        <Avatar size="md" name={post.user?.username} bg={accent} flexShrink={0} mt={1} />
        <Box flex={1}>
          <HStack spacing={2} flexWrap="wrap" mb={1}>
            <Text fontSize="sm" fontWeight="600" color={textPrimary}>{post.user?.username}</Text>
            <Box fontSize="10px" fontWeight="700" color={lv.color} bg={lv.color + '14'} px={1.5} py={0.5} borderRadius="4px">{lv.name}</Box>
            {post.user?.role === 'admin' && <Box fontSize="10px" fontWeight="700" color="#c53030" bg="#c5303014" px={1.5} py={0.5} borderRadius="4px">管理员</Box>}
            <Text fontSize="xs" color={textMuted}>#{post.post_number}</Text>
            <Text fontSize="xs" color={textMuted}>{formatDateTime(post.created_at)}</Text>
          </HStack>
          {editing ? (
            <Box>
              <Box data-color-mode={isDark ? 'dark' : 'light'} mb={2}>
                <MDEditor value={draft} onChange={(v) => setDraft(v || '')} height={220} preview="edit" />
              </Box>
              <Flex gap={2} justify="flex-end">
                <Button size="xs" variant="ghost" color={textMuted} onClick={() => setEditing(false)}>取消</Button>
                <Button size="xs" colorScheme="teal" bg={accent} color="white" _hover={{ bg: '#0a7359' }} onClick={() => { updatePostBody(post, draft); setEditing(false); }}>保存</Button>
              </Flex>
            </Box>
          ) : (
            <>
              <Box className="markdown-content" color={textPrimary} fontSize="md" lineHeight="1.75">
                <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]} components={{ code: CodeBlock }}>{post.raw}</ReactMarkdown>
              </Box>
              {actions}
            </>
          )}
        </Box>
      </HStack>
    </Box>
  );
}

const LEVEL_META: Record<number, { name: string; color: string }> = {
  0: { name: '新手', color: '#8a9099' },
  1: { name: '初级', color: '#2b6cb0' },
  2: { name: '中级', color: '#2f855a' },
  3: { name: '资深', color: '#d97706' },
};

// 提取 markdown 标题作为目录
function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { year: 'numeric', month: 'long', day: 'numeric' });
}

function formatDateTime(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}