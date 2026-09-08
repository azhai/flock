import { useMemo, useState } from 'react';
import {
  Box,
  Button,
  Flex,
  HStack,
  Icon,
  Input,
  Text,
  useColorModeValue,
} from '@chakra-ui/react';
import { FiChevronLeft, FiChevronRight, FiChevronsLeft, FiChevronsRight } from 'react-icons/fi';

interface PaginationProps {
  page: number;
  perPage: number;
  total: number;
  onChange: (page: number) => void;
  /** 当前页两侧最多显示的页号个数（默认 1） */
  siblings?: number;
}

/** 生成分页页号数组：包含首页/末页与省略号窗口。返回 string（'…' 表示省略） */
function buildPages(page: number, totalPages: number, siblings = 1): (number | string)[] {
  if (totalPages <= 7) {
    const all: number[] = [];
    for (let i = 1; i <= totalPages; i++) all.push(i);
    return all;
  }
  const start = Math.max(2, page - siblings);
  const end = Math.min(totalPages - 1, page + siblings);
  const out: (number | string)[] = [1];
  if (start > 2) out.push('…');
  for (let i = start; i <= end; i++) out.push(i);
  if (end < totalPages - 1) out.push('…');
  out.push(totalPages);
  return out;
}

/**
 * 高级分页：页号窗口 + 首页/末页 + 上一页/下一页 + 跳转输入。
 * 展示类组件，页码变化通过 onChange 由父级控制请求。
 */
export default function Pagination({ page, perPage, total, onChange, siblings = 1 }: PaginationProps) {
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const borderColor = useColorModeValue('#e8eaed', '#2d333b');
  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');
  const bg = useColorModeValue('#ffffff', '#1a1d23');

  const totalPages = useMemo(() => Math.max(1, Math.ceil(total / perPage)), [total, perPage]);
  const [jump, setJump] = useState('');

  if (totalPages <= 1) return null;

  const goto = (p: number) => {
    if (p < 1 || p > totalPages) return;
    onChange(p);
  };

  const pageBtn = (p: number | string) =>
    typeof p === 'string' ? (
      <Text key={`ell-${p}`} px={1.5} color={textSecondary} select="none">…</Text>
    ) : (
      <Button
        key={p}
        size="sm"
        minW="32px"
        px={2}
        rounded="md"
        variant="ghost"
        onClick={() => goto(p)}
        bg={p === page ? accent : 'transparent'}
        color={p === page ? '#fff' : textSecondary}
        _hover={{ bg: p === page ? accent : bgHover, color: p === page ? '#fff' : textSecondary }}
        fontWeight={p === page ? '700' : '500'}
      >
        {p}
      </Button>
    );

  const submitJump = () => {
    const n = parseInt(jump, 10);
    if (!Number.isNaN(n)) goto(n);
    setJump('');
  };

  return (
    <Flex direction={{ base: 'column', sm: 'row' }} align="center" justify="space-between" gap={3} mt={6} pt={4} borderTop="1px" borderColor={borderColor}>
      <Text fontSize="sm" color={textSecondary}>
        共 {total} 条 · 第 {page} / {totalPages} 页
      </Text>

      <HStack spacing={1}>
        <Button size="sm" variant="ghost" isDisabled={page <= 1} onClick={() => goto(1)} aria-label="首页" color={textSecondary} _hover={{ bg: bgHover }}>
          <Icon as={FiChevronsLeft} boxSize={4} />
        </Button>
        <Button size="sm" variant="ghost" isDisabled={page <= 1} onClick={() => goto(page - 1)} aria-label="上一页" color={textSecondary} _hover={{ bg: bgHover }}>
          <Icon as={FiChevronLeft} boxSize={4} />
        </Button>

        {buildPages(page, totalPages, siblings).map(pageBtn)}

        <Button size="sm" variant="ghost" isDisabled={page >= totalPages} onClick={() => goto(page + 1)} aria-label="下一页" color={textSecondary} _hover={{ bg: bgHover }}>
          <Icon as={FiChevronRight} boxSize={4} />
        </Button>
        <Button size="sm" variant="ghost" isDisabled={page >= totalPages} onClick={() => goto(totalPages)} aria-label="末页" color={textSecondary} _hover={{ bg: bgHover }}>
          <Icon as={FiChevronsRight} boxSize={4} />
        </Button>
      </HStack>

      <HStack spacing={1}>
        <Text fontSize="sm" color={textSecondary}>跳至</Text>
        <Input
          size="sm"
          w="64px"
          value={jump}
          onChange={(e) => setJump(e.target.value.replace(/\D/g, ''))}
          onKeyDown={(e) => e.key === 'Enter' && submitJump()}
          borderRadius="md"
          bg={bg}
          borderColor={borderColor}
          color={textSecondary}
          textAlign="center"
          px={1}
          aria-label="跳转页数"
        />
        <Button size="sm" variant="outline" borderColor={borderColor} color={textSecondary} onClick={submitJump} _hover={{ bg: bgHover }}>Go</Button>
      </HStack>
    </Flex>
  );
}