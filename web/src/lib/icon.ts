import m from 'mithril';

// 精简 feather 风格图标（stroke 描边），按需增删。
const PATHS: Record<string, string[]> = {
  search: ['M11 11m-7 0a7 7 0 1 0 14 0a7 7 0 1 0 -14 0', 'M21 21l-4.3 -4.3'],
  plus: ['M12 5v14M5 12h14'],
  bell: ['M18 8a6 6 0 0 0 -12 0c0 7-3 9-3 9h18s-3-2-3-9', 'M13.73 21a2 2 0 0 1 -3.46 0'],
  logout: ['M9 21H5a2 2 0 0 1 -2-2V5a2 2 0 0 1 2-2h4', 'M16 17l5-5-5-5', 'M21 12H9'],
  user: ['M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8z', 'M12 14c-4 0-7 2-7 5h14c0-3-3-5-7-5z'],
  shield: ['M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z'],
  bookmark: ['M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z'],
  thumbsUp: ['M14 9V5a3 3 0 0 0 -3-3l-4 9v11h11.28a2 2 0 0 0 2-1.7l1.38-9a2 2 0 0 0 -2-2.3z', 'M7 22H4a2 2 0 0 1 -2-2v-7a2 2 0 0 1 2-2h3'],
  atSign: ['M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0 -8 0', 'M16 8v5a3 3 0 0 0 6 0v-1a10 10 0 1 0 -3.9 7.9'],
  messageSquare: ['M21 15a2 2 0 0 1 -2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z'],
  flag: ['M4 15s1-1 4-1 5 2 8 2 4-1 4-1V3s-1 1-4 1-5-2-8-2-4 1-4 1z', 'M4 22v-7'],
  x: ['M18 6l-12 12', 'M6 6l12 12'],
  arrowUp: ['M12 19V5', 'M5 12l7-7 7 7'],
  arrowDown: ['M12 5v14', 'M5 12l7 7 7-7'],
  arrowLeft: ['M19 12H5', 'M12 19l-7-7 7-7'],
  chevronLeft: ['M15 18l-6-6 6-6'],
  chevronRight: ['M9 18l6-6-6-6'],
  chevronsLeft: ['M11 17l-5-5 5-5', 'M18 17l-5-5 5-5'],
  chevronsRight: ['M13 17l5-5-5-5', 'M6 17l5-5-5-5'],
  lock: ['M5 11h14a2 2 0 0 1 2 2v6a2 2 0 0 1 -2 2H5a2 2 0 0 1 -2-2v-6a2 2 0 0 1 2-2z', 'M8 11V7a4 4 0 0 1 8 0v4'],
  edit: ['M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z'],
  trash: ['M3 6h18', 'M19 6v14a2 2 0 0 1 -2 2H7a2 2 0 0 1 -2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2', 'M10 11v6', 'M14 11v6'],
  folder: ['M22 19a2 2 0 0 1 -2 2H4a2 2 0 0 1 -2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z'],
  tag: ['M20.59 13.41l-7.17 7.17a2 2 0 0 1 -2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z', 'M7 7m-1 0a1 1 0 1 0 2 0a1 1 0 1 0 -2 0'],
  eye: ['M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z', 'M12 12m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0'],
  eyeOff: ['M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1 -2.16 3.19M1 1l22 22', 'M9.9 9.9a3 3 0 0 0 4.2 4.2'],
  star: ['M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01z'],
  unlock: ['M5 11h14a2 2 0 0 1 2 2v6a2 2 0 0 1 -2 2H5a2 2 0 0 1 -2-2v-6a2 2 0 0 1 2-2z', 'M8 11V7a4 4 0 0 1 8 0'],
  send: ['M22 2L11 13', 'M22 2l-7 20-4-9-9-4z'],
  mail: ['M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z', 'M22 6l-10 7L2 6'],
  userCheck: ['M16 21v-2a4 4 0 0 0 -4-4H6a4 4 0 0 0 -4 4v2', 'M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8', 'M16 11l2 2 4-4'],
  smile: ['M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20z', 'M8 14s1.5 2 4 2 4-2 4-2', 'M9 9h.01', 'M15 9h.01'],
  mapPin: ['M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z', 'M12 13m-3 0a3 3 0 1 0 6 0a3 3 0 1 0 -6 0'],
};

export function Icon(attrs: any) {
  const name = attrs.name as string;
  const size = attrs.size || 18;
  const paths = PATHS[name] || [];
  return m(
    'svg',
    {
      width: size,
      height: size,
      viewBox: '0 0 24 24',
      fill: 'none',
      stroke: 'currentColor',
      'stroke-width': 2,
      'stroke-linecap': 'round',
      'stroke-linejoin': 'round',
      class: attrs.class,
    },
    paths.map((d) => m('path', { d }))
  );
}
