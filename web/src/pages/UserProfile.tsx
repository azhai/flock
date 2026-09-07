import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Box, Heading, Skeleton, Stack, Text } from '@chakra-ui/react';
import { api } from '../lib/api';

export default function UserProfile() {
  const { username } = useParams();
  const [user, setUser] = useState(null);
  const [topics, setTopics] = useState([]);

  useEffect(() => {
    if (!username) return;
    api.getUser(username).then(setUser).catch(() => {});
    api.getUserTopics(username).then((d) => setTopics(d.data || [])).catch(() => {});
  }, [username]);

  if (!user) return <Stack>{[1, 2].map((i) => <Skeleton key={i} h="40px" />)}</Stack>;

  return (
    <Stack spacing={4}>
      <Box bg="white" rounded="md" p={4} shadow="sm">
        <Heading size="lg">{user.username}</Heading>
        <Text color="gray.500">{user.full_name}</Text>
        <Text>{user.bio}</Text>
        <Text fontSize="sm" color="gray.500">
          {user.topics_count} 话题 · {user.posts_count} 帖子
        </Text>
      </Box>
      <Box bg="white" rounded="md" p={4} shadow="sm">
        <Heading size="md" mb={3}>TA的话题</Heading>
        {topics.map((t) => (
          <Box key={t.id} p={2} borderBottom="1px" borderColor="gray.100">
            <Heading size="sm">{t.title}</Heading>
            <Text fontSize="sm" color="gray.500">{t.posts_count} 回复 · {t.created_at}</Text>
          </Box>
        ))}
      </Box>
    </Stack>
  );
}