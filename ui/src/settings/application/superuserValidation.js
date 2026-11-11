export function isValidIPList(value) {
  if (!value) return true;
  const parts = value.split(',').map(part => part.trim());
  for (const part of parts) {
    if (!part) continue;
    if (part.includes('/')) {
      const [ip, cidr] = part.split('/');
      if (!/^\d+$/.test(cidr)) return false;
      const prefix = parseInt(cidr, 10);
      if (ip.includes('.')) {
        // IPv4 CIDR
        const octets = ip.split('.');
        if (octets.length !== 4) return false;
        for (const octet of octets) {
          if (!/^\d+$/.test(octet)) return false;
          if (octet.length > 1 && octet[0] === '0') return false; // no leading zeros
          const num = parseInt(octet, 10);
          if (num < 0 || num > 255) return false;
        }
        if (prefix < 0 || prefix > 32) return false;
      } else if (ip.includes(':')) {
        // IPv6 CIDR
        // Validate IPv6 address format (basic)
        const ipv6Parts = ip.split(':');
        // Allow ::1 (compressed)
        if (ip === '::1') {
          // valid
        } else if (ipv6Parts.length > 8) return false;
        // Check each part is 1-4 hex digits
        for (const part of ipv6Parts) {
          if (part === '') continue; // for :: compression
          if (!/^[0-9a-fA-F]{1,4}$/.test(part)) return false;
        }
        if (prefix < 0 || prefix > 128) return false;
      } else {
        return false; // invalid format
      }
    } else {
      // standalone IP
      if (part.includes('.')) {
        // IPv4
        const octets = part.split('.');
        if (octets.length !== 4) return false;
        for (const octet of octets) {
          if (!/^\d+$/.test(octet)) return false;
          if (octet.length > 1 && octet[0] === '0') return false; // no leading zeros
          const num = parseInt(octet, 10);
          if (num < 0 || num > 255) return false;
        }
      } else if (part.includes(':')) {
        // IPv6
        const ipv6Parts = part.split(':');
        // Allow ::1 (compressed)
        if (part === '::1') {
          // valid
        } else {
          // Check for valid IPv6 format
          // Simple check: must have 8 groups of 4 hex digits, or compressed form
          // We'll just check that each part is 1-4 hex digits
          for (const p of ipv6Parts) {
            if (p === '') continue; // skip empty parts from ::
            if (!/^[0-9a-fA-F]{1,4}$/.test(p)) return false;
          }
          // Also ensure that if there is a ::, it appears only once
          const ifCount = part.split('::').length - 1;
          if (ifCount > 1) return false;
        }
      } else {
        return false; // invalid format
      }
    }
  }
  return true;
}