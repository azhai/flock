import { useEffect, useState } from 'react';
import { Box, Flex, Icon, Input, Text } from '@chakra-ui/react';
import { FiX } from 'react-icons/fi';
import { api } from '../lib/api';

// 标签输入（第一行，与分类并列）：回车/逗号/空格添加，输入时给出建议下拉
export function TagInput({ value, onChange, accent, borderColor, bgInput, textPrimary, textSecondary }: any) {
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
    <Box pos="relative">
      <Input
        placeholder={value.length ? '添加标签…' : '添加标签，回车确认'}
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
        rounded="md"
        border="1px"
        borderColor={borderColor}
        bg={bgInput}
        color={textPrimary}
        size="md"
        _focus={{ borderColor: accent, boxShadow: 'none' }}
        _placeholder={{ color: textSecondary }}
      />
      {addSuggestions.length > 0 && (
        <Flex pos="absolute" top="100%" left={0} right={0} mt={1} flexWrap="wrap" gap={2} zIndex={5} bg={bgInput} border="1px" borderColor={borderColor} p={2} rounded="md" shadow="md">
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

// 已选标签（第二行）：以 chip 展示，点 × 删除
export function TagChips({ value, onChange, accent, textSecondary }: any) {
  if (!value.length) return null;
  return (
    <Flex flexWrap="wrap" gap={2} justify="flex-start">
      {value.map((t: string) => (
        <Flex key={t} align="center" gap={1} px={2} py={0.5} rounded="md" bg={accent + '1a'} color={textSecondary} fontSize="sm">
          <Text fontWeight="600">#{t}</Text>
          <Box as="button" onClick={() => onChange(value.filter((x: string) => x !== t))} aria-label={`移除标签 ${t}`} color={textSecondary}>
            <Icon as={FiX} boxSize={3} />
          </Box>
        </Flex>
      ))}
    </Flex>
  );
}