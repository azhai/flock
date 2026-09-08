import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Button,
  Flex,
  Heading,
  Icon,
  Input,
  Popover,
  PopoverBody,
  PopoverContent,
  PopoverTrigger,
  Select,
  Stack,
  Text,
  useColorModeValue,
  useToast,
} from '@chakra-ui/react';
import { FiSend, FiEdit3, FiLogIn, FiX, FiSmile } from 'react-icons/fi';
import MDEditor from '@uiw/react-md-editor/nohighlight';
import '@uiw/react-md-editor/markdown-editor.css';
import CodeBlock from '../components/CodeBlock';
import { ensureSyntaxHighlight } from '../lib/highlightProvider';
import { api } from '../lib/api';
import { useStore } from '../lib/store';

export default function NewTopic() {
  const { user } = useStore();
  const nav = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [raw, setRaw] = useState('');
  const [cats, setCats] = useState<any[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [isDark] = useState(() => false); // 单一浅色主题，编辑器始终 light

  const bgMain = useColorModeValue('#ffffff', '#1a1d23');
  const bgInput = useColorModeValue('#f7f8fa', '#21252c');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  useEffect(() => {
    api.listCategories().then(setCats).catch(() => {});
    ensureSyntaxHighlight().catch(() => {}); // 懒加载语法高亮器
  }, []);

  const submit = () => {
    if (!title.trim()) {
      toast({ title: '请填写标题', status: 'warning' });
      return;
    }
    if (!raw.trim()) {
      toast({ title: '请填写内容', status: 'warning' });
      return;
    }
    setSubmitting(true);
    api
      .createTopic({
        title: title.trim(),
        raw: raw.trim(),
        category_id: categoryId ? Number(categoryId) : null,
        tags,
      })
      .then((t: any) => {
        toast({ title: '发布成功', status: 'success' });
        nav(`/topics/${t.id || t.topic?.id}`);
      })
      .catch((e: any) => toast({ title: e || '发布失败', status: 'error' }))
      .finally(() => setSubmitting(false));
  };

  // 未登录：引导登录（登录后自动回到本页）
  if (!user) {
    return (
      <Flex minH="calc(100vh - 56px)" align="center" justify="center" bg={bgMain} px={4}>
        <Box bg={bgMain} rounded="lg" p={8} shadow="sm" textAlign="center" maxW="sm" w="full">
          <Icon as={FiEdit3} boxSize={10} color={accent} mb={4} />
          <Heading size="md" color={textPrimary} mb={2}>登录后即可发帖</Heading>
          <Text color={textSecondary} mb={6} fontSize="sm">发表话题需要先登录账号</Text>
          <Button
            as="a"
            href="/login"
            bg={accent} color="white" _hover={{ bg: '#0a7359' }}
            leftIcon={<FiLogIn />}
            rounded="md"
            onClick={(e: any) => {
              e.preventDefault();
              nav('/login', { state: { from: '/new' } });
            }}
          >
            去登录
          </Button>
        </Box>
      </Flex>
    );
  }

  return (
    <Box bg={bgMain} minH="calc(100vh - 56px)" py={8} px={4}>
      <Box maxW="680px" mx="auto">
        <Heading size="md" color={textPrimary} fontWeight="700" mb={1}>发布新话题</Heading>
        <Text color={textSecondary} fontSize="sm" mb={6}>内容支持 Markdown 语法，可为话题添加标签</Text>

        <Stack spacing={4}>
          <Box bg={bgMain} rounded="lg" shadow="sm" border="1px" borderColor={borderColor} p={5}>
            <Stack spacing={4}>
              <Input
                placeholder="标题：一句话说清你的话题"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                size="md"
                rounded="md"
                border="1px"
                borderColor={borderColor}
                bg={bgInput}
                color={textPrimary}
                fontSize="md"
                fontWeight="600"
                _focus={{ borderColor: accent, boxShadow: 'none' }}
                _placeholder={{ color: textMuted, fontWeight: '400' }}
              />

              <Select
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                rounded="md"
                border="1px"
                borderColor={borderColor}
                bg={bgInput}
                color={textPrimary}
                size="md"
                _focus={{ borderColor: accent, boxShadow: 'none' }}
              >
                <option value="">选择分类（可选）</option>
                {cats.map((c: any) => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </Select>

              <TagInput value={tags} onChange={setTags} accent={accent} borderColor={borderColor} bgInput={bgInput} textPrimary={textPrimary} textSecondary={textSecondary} />
            </Stack>
          </Box>

          <Box data-color-mode={isDark ? 'dark' : 'light'}>
            <Flex justify="flex-end" mb={1}>
              <Popover placement="top-end">
                <PopoverTrigger>
                  <Button size="xs" variant="ghost" leftIcon={<Icon as={FiSmile} boxSize={3} />} color={textSecondary} _hover={{ bg: bgInput, color: textPrimary }}>表情</Button>
                </PopoverTrigger>
                <PopoverContent w="320px" bg={bgMain} borderColor={borderColor} mt={1}>
                  <PopoverBody>
                    <Flex flexWrap="wrap" gap={0.5}>
                      {EMOJIS.map((e) => (
                        <Box
                          key={e}
                          as="button"
                          onClick={() => setRaw((r) => r + e)}
                          fontSize="20px"
                          lineHeight="1.6"
                          w="36px"
                          h="36px"
                          textAlign="center"
                          borderRadius="md"
                          _hover={{ bg: bgInput }}
                        >
                          {e}
                        </Box>
                      ))}
                    </Flex>
                  </PopoverBody>
                </PopoverContent>
              </Popover>
            </Flex>
            <MDEditor
              value={raw}
              onChange={(v) => setRaw(v || '')}
              height={340}
              preview="edit"
              previewOptions={{ components: { code: CodeBlock } }}
              textareaProps={{ placeholder: '正文内容…\n\n支持 **粗体**、*斜体*、`代码`、列表、引用等 Markdown 语法' }}
            />
          </Box>

          <Flex justify="space-between" align="center" gap={2}>
            <Text fontSize="xs" color={textMuted}>{raw.length} 字 · 单回车自动换行</Text>
            <Flex justify="flex-end" gap={2}>
              <Button variant="ghost" size="sm" onClick={() => nav(-1)} color={textSecondary}>
                取消
              </Button>
              <Button
                bg={accent} color="white" _hover={{ bg: '#0a7359' }}
                size="sm"
                onClick={submit}
                isLoading={submitting}
                rightIcon={<FiSend />}
                rounded="md"
                px={5}
              >
                发布话题
              </Button>
            </Flex>
          </Flex>
        </Stack>
      </Box>
    </Box>
  );
}

// 标签输入：回车/逗号添加为 chip，可点击删除
function TagInput({ value, onChange, accent, borderColor, bgInput, textPrimary, textSecondary }: any) {
  const [draft, setDraft] = useState('');
  const [suggestions, setSuggestions] = useState<any[]>([]);

  useEffect(() => {
    api.listTags().then((arr: any) => setSuggestions(Array.isArray(arr) ? arr : [])).catch(() => {});
  }, []);

  const commit = (name: string) => {
    const n = name.trim().replace(/^#/, '');
    if (!n) return;
    if (!value.some((t: string) => t.toLowerCase() === n.toLowerCase()) && value.length < 8) {
      onChange([...value, n]);
    }
    setDraft('');
  };

  const addSuggestions = draft
    ? suggestions.filter((s) => s.name.includes(draft.toLowerCase()) && !value.includes(s.name)).slice(0, 5)
    : [];

  return (
    <Box>
      <Flex flexWrap="wrap" gap={2} align="center" minH="40px" p={2} rounded="md" border="1px" borderColor={borderColor} bg={bgInput}>
        {value.map((t: string) => (
          <Flex key={t} align="center" gap={1} px={2} py={0.5} rounded="md" bg={accent + '1a'} color={textSecondary} fontSize="sm">
            <Text fontWeight="600">#{t}</Text>
            <Box as="button" onClick={() => onChange(value.filter((x: string) => x !== t))} aria-label={`移除标签 ${t}`} color={textSecondary}>
              <Icon as={FiX} boxSize={3} />
            </Box>
          </Flex>
        ))}
        <Box flex={1}>
          <Input
            variant="unstyled"
            placeholder={value.length ? '添加更多标签…' : '添加标签，回车确认'}
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e: any) => {
              if (e.key === 'Enter' || e.key === ',' || e.key === ' ') {
                e.preventDefault();
                commit(draft);
              } else if (e.key === 'Backspace' && !draft && value.length) {
                onChange(value.slice(0, -1));
              }
            }}
            color={textPrimary}
            fontSize="sm"
            _placeholder={{ color: textSecondary }}
          />
        </Box>
      </Flex>
      {addSuggestions.length > 0 && (
        <Flex flexWrap="wrap" gap={2} mt={2}>
          {addSuggestions.map((s: any) => (
            <Box key={s.id} as="button" onClick={() => { commit(s.name); }} bg="transparent" border="1px" borderColor={borderColor} color={textSecondary} px={2} py={0.5} rounded="md" fontSize="xs" _hover={{ bg: accent + '1a', color: textPrimary }}>
              #{s.name} · {s.topics_count}
            </Box>
          ))}
        </Flex>
      )}
    </Box>
  );
}

const EMOJIS = [
  '😀', '😄', '😂', '🤣', '😊', '😍', '🥰', '😎', '🤔', '😅',
  '🙌', '👏', '👍', '👎', '🙏', '💪', '🤝', '👋', '🎉', '🔥',
  '💯', '✅', '❌', '❤️', '💚', '💡', '❗', '❓', '🚀', '📌',
  '🏆', '📢', '👀', '💬', '🗃️', '⚡', '☁️', '🐛', '🧩', '📝',
];