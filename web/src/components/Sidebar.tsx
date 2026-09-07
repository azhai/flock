import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import {
  Box,
  Flex,
  Icon,
  Text,
  useColorModeValue,
  VStack,
} from '@chakra-ui/react';
import { FiMessageSquare } from 'react-icons/fi';
import { api } from '../lib/api';

// 左栏：分类 + 热门标签。
// 选中状态从 URL (?cat=<slug>&tag=<slug>) 读取，点击任意项跳回首页并按所选筛选，
// 因此从话题详情返回首页时，选中项仍保持不变。
export default function Sidebar() {
  const [searchParams] = useSearchParams();
  const nav = useNavigate();
  const [cats, setCats] = useState<any[]>([]);
  const [tags, setTags] = useState<any[]>([]);

  const activeCat = searchParams.get('cat');
  const activeTag = searchParams.get('tag');

  useEffect(() => {
    api.listCategories().then(setCats).catch(() => {});
    api.listTags().then((arr: any) => setTags(Array.isArray(arr) ? arr : [])).catch(() => {});
  }, []);

  const bgHover = useColorModeValue('#f3f5f7', '#272c34');
  const textPrimary = useColorModeValue('#1f2329', '#e6e8eb');
  const textSecondary = useColorModeValue('#5f6571', '#a0a6ad');
  const textMuted = useColorModeValue('#8a9099', '#6c727a');
  const accent = useColorModeValue('#0d8a6f', '#3db99a');

  const selectCat = (slug: string | null) => {
    const sp = new URLSearchParams();
    if (slug) sp.set('cat', slug);
    if (activeTag) sp.set('tag', activeTag);
    const qs = sp.toString();
    nav(qs ? `/?${qs}` : '/');
  };

  const selectTag = (slug: string | null) => {
    const sp = new URLSearchParams();
    if (activeCat) sp.set('cat', activeCat);
    if (slug) sp.set('tag', slug);
    const qs = sp.toString();
    nav(qs ? `/?${qs}` : '/');
  };

  return (
    <Box py={4} px={3}>
      <Text fontSize="xs" color={textMuted} fontWeight="700" textTransform="uppercase" letterSpacing="0.5px" px={2} mb={2}>
        分类
      </Text>
      <VStack align="stretch" spacing={0.5}>
        <CategoryItem
          label="全部"
          colorDot={null}
          active={!activeCat && !activeTag}
          onClick={() => selectCat(null)}
          bgHover={bgHover}
          accent={accent}
          textPrimary={textPrimary}
          textSecondary={textSecondary}
        />
        {cats.map((c: any) => (
          <CategoryItem
            key={c.id}
            label={c.name}
            colorDot={c.color}
            active={activeCat === c.slug}
            onClick={() => selectCat(c.slug)}
            bgHover={bgHover}
            accent={accent}
            textPrimary={textPrimary}
            textSecondary={textSecondary}
          />
        ))}
      </VStack>

      {tags.length > 0 && (
        <>
          <Text fontSize="xs" color={textMuted} fontWeight="700" textTransform="uppercase" letterSpacing="0.5px" px={2} mb={2} mt={5}>
            热门标签
          </Text>
          <Flex flexWrap="wrap" gap={1.5} px={1}>
            {tags.slice(0, 12).map((tg: any) => {
              const active = activeTag === tg.slug;
              return (
                <Box
                  as="button"
                  key={tg.id}
                  onClick={() => selectTag(active ? null : tg.slug)}
                  fontSize="sm"
                  fontWeight={active ? '700' : '500'}
                  color={active ? accent : textSecondary}
                  bg={active ? accent + '14' : 'transparent'}
                  border="1px"
                  borderColor={active ? accent : 'transparent'}
                  px={2}
                  py={0.5}
                  borderRadius="full"
                  _hover={{ bg: accent + '14', color: active ? accent : textPrimary }}
                >
                  #{tg.name}
                  <Text as="span" color={textMuted} fontWeight="400">· {tg.topics_count}</Text>
                </Box>
              );
            })}
          </Flex>
        </>
      )}
    </Box>
  );
}

function CategoryItem({ label, colorDot, active, onClick, bgHover, accent, textPrimary, textSecondary }: any) {
  return (
    <Box
      as="button"
      onClick={onClick}
      display="flex"
      alignItems="center"
      gap={2}
      px={2}
      py={1.5}
      pl={3}
      rounded="md"
      w="full"
      textAlign="left"
      bg={active ? accent + '1a' : 'transparent'}
      color={active ? accent : textSecondary}
      fontWeight={active ? '700' : '400'}
      fontSize="md"
      // 选中态：左边直角，右边一个很大的钝角箭头（clip-path 标签形状）
      style={active ? { clipPath: 'polygon(0 0, calc(100% - 16px) 0, 100% 50%, calc(100% - 16px) 100%, 0 100%)' } : undefined}
      _hover={active ? { bg: accent + '1a' } : { bg: bgHover, color: textPrimary }}
      cursor="pointer"
      transition="all 0.15s ease"
    >
      {colorDot ? <Box w={2} h={2} rounded="full" bg={colorDot} flexShrink={0} /> : <Icon as={FiMessageSquare} boxSize={4} flexShrink={0} />}
      <Text isTruncated px={0}>{label}</Text>
    </Box>
  );
}