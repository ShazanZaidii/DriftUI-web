import React from 'react';

export function Home() {
  return (
    <div style={{ display: 'flex', flexDirection: 'column' }}>
      <span>Engine Booted.</span>
      <span>Routing to custom components...</span>
      <ProfileCard />
    </div>
  );
}
export function ProfileCard() {
  return (
    <div style={{ display: 'grid', gridTemplateColumns: '1fr', gridTemplateRows: '1fr' }}>
      <div style={{ gridArea: '1 / 1 / 2 / 2' }}>
    <div style={{ display: 'flex', flexDirection: 'row' }}>
      <span>Name: Drift Dev</span>
      <span>Status: Online</span>
    </div>
      </div>
    </div>
  );
}
